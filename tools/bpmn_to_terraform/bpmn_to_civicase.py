#!/usr/bin/env python3
"""
bpmn_to_civicase.py — Convert a BPMN process file to a civicrm_case_type Terraform resource.

BPMN → CiviCase mapping:
  userTask / task / manualTask / serviceTask / scriptTask
      → activityType in definition.activityTypes
      → activityType in timeline activitySet (with sequential offset)
  parallelGateway (fork+join pairs)
      → mapped as comment in .tf; activities inside parallel branch get
        a note that they run concurrently (CiviCase has no native parallel)
  exclusiveGateway / inclusiveGateway
      → same timeline position; condition text preserved as description comment
  startEvent
      → "Open Case" activity (max_instances = 1)
  endEvent
      → "Change Case Status" activity at end of timeline
  process.name / process.id
      → resource name / title
  lanes (laneSet / lane)
      → caseRoles: each lane becomes a role; tasks in lane get role annotation
  boundaryEvent (timer)
      → reference_offset on the closest following activity
  dataObject / dataStore
      → not mappable; emitted as comment suggestion in output
  subProcess / callActivity
      → flattened; tasks inside are inlined with a comment
  textAnnotation / association
      → text preserved as description on nearest task

Usage:
  python bpmn_to_civicase.py input.bpmn [--output output.tf] [--weight 1]
  python bpmn_to_civicase.py input.bpmn --print-mapping   # show mapping report
"""

import argparse
import json
import re
import sys
import textwrap
import xml.etree.ElementTree as ET
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional


# ─── BPMN namespace aliases ──────────────────────────────────────────────────

NS = {
    "bpmn":   "http://www.omg.org/spec/BPMN/20100524/MODEL",
    "bpmn2":  "http://www.omg.org/spec/BPMN/20100524/MODEL",
    "omgdc":  "http://www.omg.org/spec/DD/20100524/DC",
    "omgdi":  "http://www.omg.org/spec/DD/20100524/DI",
    "activiti": "http://activiti.org/bpmn",
    "camunda":  "http://camunda.org/schema/1.0/bpmn",
}

TASK_TAGS = {
    "userTask", "task", "manualTask", "serviceTask",
    "scriptTask", "sendTask", "receiveTask", "businessRuleTask",
}

GATEWAY_TAGS = {
    "parallelGateway", "exclusiveGateway", "inclusiveGateway",
    "eventBasedGateway", "complexGateway",
}

# Tags we silently skip (layout / diagram info)
SKIP_TAGS = {
    "BPMNDiagram", "BPMNPlane", "BPMNShape", "BPMNEdge",
    "Bounds", "waypoint",
}


# ─── Data model ──────────────────────────────────────────────────────────────

@dataclass
class BpmnTask:
    id: str
    name: str
    tag: str
    lane: Optional[str] = None
    description: Optional[str] = None
    parallel_group: Optional[str] = None  # fork gateway id
    timer_offset: Optional[int] = None    # days from boundary timer


@dataclass
class BpmnGateway:
    id: str
    name: str
    tag: str
    incoming: list = field(default_factory=list)
    outgoing: list = field(default_factory=list)


@dataclass
class BpmnLane:
    id: str
    name: str
    flow_node_refs: list = field(default_factory=list)


@dataclass
class BpmnProcess:
    id: str
    name: str
    tasks: list = field(default_factory=list)           # BpmnTask list in topo order
    gateways: dict = field(default_factory=dict)        # id → BpmnGateway
    lanes: list = field(default_factory=list)           # BpmnLane list
    sequence_flows: dict = field(default_factory=dict)  # id → {source, target}
    text_annotations: dict = field(default_factory=dict)  # id → text
    associations: list = field(default_factory=list)    # [(src, tgt)]
    warnings: list = field(default_factory=list)


# ─── Helpers ─────────────────────────────────────────────────────────────────

def _local(tag: str) -> str:
    """Strip namespace prefix from an XML tag."""
    return tag.split("}")[-1] if "}" in tag else tag


def _attr(el: ET.Element, name: str, default: str = "") -> str:
    return el.attrib.get(name, default)


def _slugify(text: str) -> str:
    """Convert a display name to a Terraform/CiviCRM machine name."""
    slug = text.lower()
    slug = re.sub(r"[^a-z0-9]+", "_", slug)
    slug = slug.strip("_")
    return slug or "activity"


def _tf_string(value: str) -> str:
    """Escape a string for use in HCL."""
    return value.replace("\\", "\\\\").replace('"', '\\"')


# ─── Parser ──────────────────────────────────────────────────────────────────

def parse_bpmn(path: Path) -> BpmnProcess:
    tree = ET.parse(path)
    root = tree.getroot()

    # Register all namespaces so ET resolves them
    for prefix, uri in NS.items():
        ET.register_namespace(prefix, uri)

    def find_process(el: ET.Element) -> Optional[ET.Element]:
        tag = _local(el.tag)
        if tag == "process":
            return el
        for child in el:
            result = find_process(child)
            if result is not None:
                return result
        return None

    proc_el = find_process(root)
    if proc_el is None:
        sys.exit("ERROR: No <process> element found in BPMN file.")

    proc = BpmnProcess(
        id=_attr(proc_el, "id", "process"),
        name=_attr(proc_el, "name", "Unnamed Process"),
    )

    # ── Build lookup maps ──────────────────────────────────────────────────

    # sequence flows: id → {source, target}
    # node ids → tag name
    node_tag: dict[str, str] = {}
    # adjacency: node id → list of successor node ids
    successors: dict[str, list[str]] = {}
    predecessors: dict[str, list[str]] = {}
    all_elements: dict[str, ET.Element] = {}

    # Text annotations
    for el in proc_el.iter():
        tag = _local(el.tag)
        eid = _attr(el, "id")
        if eid:
            all_elements[eid] = el
        if tag == "textAnnotation":
            text_el = el.find(".//{*}text")
            proc.text_annotations[eid] = text_el.text.strip() if text_el is not None and text_el.text else ""
        if tag == "association":
            src = _attr(el, "sourceRef")
            tgt = _attr(el, "targetRef")
            if src and tgt:
                proc.associations.append((src, tgt))
        if tag == "sequenceFlow":
            src = _attr(el, "sourceRef")
            tgt = _attr(el, "targetRef")
            flow_id = _attr(el, "id")
            proc.sequence_flows[flow_id] = {"source": src, "target": tgt}
            successors.setdefault(src, []).append(tgt)
            predecessors.setdefault(tgt, []).append(src)

    # ── Lanes ──────────────────────────────────────────────────────────────
    lane_of_node: dict[str, str] = {}
    for lane_el in proc_el.iter():
        if _local(lane_el.tag) != "lane":
            continue
        lane = BpmnLane(
            id=_attr(lane_el, "id"),
            name=_attr(lane_el, "name", "Unnamed Lane"),
        )
        for ref_el in lane_el:
            if _local(ref_el.tag) == "flowNodeRef":
                node_id = (ref_el.text or "").strip()
                lane.flow_node_refs.append(node_id)
                lane_of_node[node_id] = lane.name
        proc.lanes.append(lane)

    # ── Collect raw nodes ─────────────────────────────────────────────────
    tasks_by_id: dict[str, BpmnTask] = {}
    gateways_by_id: dict[str, BpmnGateway] = {}

    for el in proc_el:
        tag = _local(el.tag)
        eid = _attr(el, "id")
        if not eid:
            continue
        node_tag[eid] = tag
        successors.setdefault(eid, [])
        predecessors.setdefault(eid, [])

        if tag in TASK_TAGS:
            task = BpmnTask(
                id=eid,
                name=_attr(el, "name", eid),
                tag=tag,
                lane=lane_of_node.get(eid),
            )
            tasks_by_id[eid] = task

        elif tag in GATEWAY_TAGS:
            gw = BpmnGateway(
                id=eid,
                name=_attr(el, "name", ""),
                tag=tag,
            )
            gateways_by_id[eid] = gw

        elif tag == "subProcess":
            proc.warnings.append(
                f"subProcess '{_attr(el, 'name', eid)}' flattened — nested tasks inlined."
            )
            # Inline sub-process tasks
            for sub_el in el.iter():
                sub_tag = _local(sub_el.tag)
                sub_id = _attr(sub_el, "id")
                if sub_tag in TASK_TAGS and sub_id:
                    task = BpmnTask(
                        id=sub_id,
                        name=_attr(sub_el, "name", sub_id),
                        tag=sub_tag,
                        lane=lane_of_node.get(sub_id),
                        description=f"(from subProcess: {_attr(el, 'name', eid)})",
                    )
                    tasks_by_id[sub_id] = task

        elif tag == "callActivity":
            proc.warnings.append(
                f"callActivity '{_attr(el, 'name', eid)}' treated as userTask."
            )
            task = BpmnTask(
                id=eid,
                name=_attr(el, "name", eid),
                tag="userTask",
                lane=lane_of_node.get(eid),
                description="(originally callActivity)",
            )
            tasks_by_id[eid] = task

        elif tag in ("dataObjectReference", "dataStoreReference", "dataObject", "dataStore"):
            proc.warnings.append(
                f"Data element '{_attr(el, 'name', eid)}' ({tag}) cannot be mapped to CiviCase. "
                "Consider storing this as a custom field on the case."
            )

        elif tag == "boundaryEvent":
            # Look for timer definition
            attached_to = _attr(el, "attachedToRef")
            for sub in el.iter():
                if _local(sub.tag) == "timeDuration":
                    # Parse ISO 8601 duration like P7D → 7 days
                    duration_text = (sub.text or "").strip()
                    days = _parse_iso_duration_days(duration_text)
                    if days is not None and attached_to in tasks_by_id:
                        tasks_by_id[attached_to].timer_offset = days

        elif tag == "intermediateCatchEvent":
            proc.warnings.append(
                f"intermediateCatchEvent '{_attr(el, 'name', eid)}' mapped to a Follow Up activity."
            )
            task = BpmnTask(
                id=eid,
                name=_attr(el, "name", eid) or "Intermediate Event",
                tag="userTask",
                description="(originally intermediateCatchEvent)",
            )
            tasks_by_id[eid] = task

    # ── Associate text annotations to tasks ───────────────────────────────
    for (src, tgt) in proc.associations:
        ann_text = proc.text_annotations.get(src) or proc.text_annotations.get(tgt)
        task_id = tgt if tgt in tasks_by_id else (src if src in tasks_by_id else None)
        if ann_text and task_id:
            existing = tasks_by_id[task_id].description or ""
            tasks_by_id[task_id].description = (existing + " " + ann_text).strip()

    # ── Identify parallel groups (fork→join pairs) ────────────────────────
    # For each parallelGateway fork, find all tasks reachable before the join
    parallel_group_of: dict[str, str] = {}  # task_id → fork_gateway_id
    for gw_id, gw in gateways_by_id.items():
        if gw.tag != "parallelGateway":
            continue
        if len(successors.get(gw_id, [])) > 1:
            # This is a fork — BFS to find parallel tasks
            visited = set()
            queue = list(successors[gw_id])
            while queue:
                nid = queue.pop(0)
                if nid in visited:
                    continue
                visited.add(nid)
                ntag = node_tag.get(nid, "")
                if ntag in TASK_TAGS:
                    parallel_group_of[nid] = gw_id
                elif ntag in GATEWAY_TAGS and nid != gw_id:
                    # Stop at join gateway
                    if gateways_by_id[nid].tag == "parallelGateway" and len(predecessors.get(nid, [])) > 1:
                        continue
                queue.extend(successors.get(nid, []))

    for task_id, fork_id in parallel_group_of.items():
        if task_id in tasks_by_id:
            tasks_by_id[task_id].parallel_group = fork_id

    # ── Topological sort of tasks ─────────────────────────────────────────
    ordered_tasks = _toposort_tasks(tasks_by_id, successors, node_tag)

    proc.tasks = ordered_tasks
    proc.gateways = gateways_by_id
    return proc


def _parse_iso_duration_days(text: str) -> Optional[int]:
    """Parse simplified ISO 8601 duration (P7D, PT48H, P1W) → integer days."""
    m = re.match(r"P(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?)?", text or "")
    if not m:
        return None
    weeks = int(m.group(1) or 0)
    days = int(m.group(2) or 0)
    hours = int(m.group(3) or 0)
    total = weeks * 7 + days + hours // 24
    return total if total > 0 else None


def _toposort_tasks(
    tasks_by_id: dict,
    successors: dict,
    node_tag: dict,
) -> list:
    """Return tasks in execution order via BFS from start events."""
    # We do a simple BFS over the full graph, collecting tasks in encounter order.
    visited_nodes: set = set()
    ordered: list = []

    # Find start nodes (no predecessors in successors graph)
    all_nodes = set(node_tag.keys())
    has_incoming = set()
    for src, targets in successors.items():
        has_incoming.update(targets)
    start_nodes = [n for n in all_nodes if n not in has_incoming]
    if not start_nodes:
        start_nodes = list(all_nodes)[:1]

    queue = list(start_nodes)
    while queue:
        nid = queue.pop(0)
        if nid in visited_nodes:
            continue
        visited_nodes.add(nid)
        if nid in tasks_by_id:
            ordered.append(tasks_by_id[nid])
        queue.extend(successors.get(nid, []))

    # Append any tasks not reached (disconnected)
    for tid, task in tasks_by_id.items():
        if tid not in visited_nodes:
            ordered.append(task)

    return ordered


# ─── Transformer: BpmnProcess → CiviCase definition ──────────────────────────

def build_civicase_definition(proc: BpmnProcess) -> dict:
    """Build the CiviCase definition JSON structure from a parsed BPMN process."""

    # ── Activity Types (unique names) ──────────────────────────────────────
    activity_type_names: list[str] = ["Open Case"]
    seen_names: set = {"open case"}

    for task in proc.tasks:
        norm = task.name.strip()
        if norm.lower() not in seen_names:
            seen_names.add(norm.lower())
            activity_type_names.append(norm)

    # Always include Change Case Status for end event
    if "change case status" not in seen_names:
        activity_type_names.append("Change Case Status")

    activity_types = []
    for name in activity_type_names:
        entry: dict = {"name": name}
        if name == "Open Case":
            entry["max_instances"] = "1"
        activity_types.append(entry)

    # ── Timeline activity set ──────────────────────────────────────────────
    # Each task becomes a timeline entry with a reference_offset.
    # Parallel groups get the same reference_offset (same "day").
    timeline_entries = []

    # Start with Open Case
    timeline_entries.append({
        "name": "Open Case",
        "status": "Completed",
    })

    offset_days = 0
    last_parallel_fork: Optional[str] = None
    parallel_offset: dict[str, int] = {}  # fork_id → offset when fork started

    for task in proc.tasks:
        name = task.name.strip()

        if task.parallel_group:
            fork_id = task.parallel_group
            if fork_id not in parallel_offset:
                offset_days += 1
                parallel_offset[fork_id] = offset_days
            entry_offset = parallel_offset[fork_id]
        else:
            if last_parallel_fork and last_parallel_fork in parallel_offset:
                # After join, advance past the parallel block
                offset_days = parallel_offset[last_parallel_fork] + 1
                last_parallel_fork = None
            else:
                offset_days += 1
            entry_offset = offset_days

        if task.parallel_group and task.parallel_group != last_parallel_fork:
            last_parallel_fork = task.parallel_group

        entry: dict = {
            "name": name,
            "reference_activity": "Open Case",
            "reference_offset": str(entry_offset),
            "reference_select": "newest",
        }

        if task.timer_offset is not None:
            entry["reference_offset"] = str(task.timer_offset)

        timeline_entries.append(entry)

    # End: Change Case Status
    timeline_entries.append({
        "name": "Change Case Status",
        "reference_activity": "Open Case",
        "reference_offset": str(offset_days + 1),
        "reference_select": "newest",
    })

    activity_sets = [
        {
            "name": "standard_timeline",
            "label": "Standard Timeline",
            "timeline": 1,
            "activityTypes": timeline_entries,
        }
    ]

    # ── Case Roles from lanes ──────────────────────────────────────────────
    case_roles = []
    if proc.lanes:
        for i, lane in enumerate(proc.lanes):
            role: dict = {"name": lane.name}
            if i == 0:
                role["creator"] = "1"
                role["manager"] = "1"
            case_roles.append(role)
    else:
        case_roles.append({"name": "Case Coordinator", "creator": "1", "manager": "1"})

    definition = {
        "activityTypes": activity_types,
        "activitySets": activity_sets,
        "caseRoles": case_roles,
    }

    return definition


# ─── HCL renderer ────────────────────────────────────────────────────────────

def render_terraform(proc: BpmnProcess, definition: dict, weight: int) -> str:
    resource_name = _slugify(proc.id)
    title = proc.name or proc.id
    description = f"Case type generated from BPMN process '{proc.id}'"

    # ── Mapping commentary ────────────────────────────────────────────────
    comment_lines = ["# BPMN → CiviCase mapping notes:"]

    # Parallel gateways
    parallel_gws = [gw for gw in proc.gateways.values() if gw.tag == "parallelGateway"]
    if parallel_gws:
        comment_lines.append("#")
        comment_lines.append(
            "# PARALLEL GATEWAYS detected — CiviCase has no native parallel execution."
        )
        comment_lines.append(
            "# Parallel tasks are assigned the same timeline offset (same 'day')."
        )
        comment_lines.append(
            "# To enforce parallel completion, consider using CiviRules or a custom extension."
        )

    # Exclusive/inclusive gateways
    exc_gws = [gw for gw in proc.gateways.values() if gw.tag in ("exclusiveGateway", "inclusiveGateway")]
    if exc_gws:
        comment_lines.append("#")
        comment_lines.append(
            "# CONDITIONAL GATEWAYS detected — CiviCase timelines are linear."
        )
        comment_lines.append(
            "# Conditions cannot be modelled natively. Options:"
        )
        comment_lines.append(
            "#   • Use CiviRules to trigger activities based on case status/custom fields."
        )
        comment_lines.append(
            "#   • Add multiple activity sets (one per branch) and choose manually."
        )

    # Warnings from parser
    for w in proc.warnings:
        comment_lines.append(f"# WARNING: {w}")

    # Unmapped elements suggestions
    comment_lines.append("#")
    comment_lines.append(
        "# Elements with no direct CiviCase equivalent:"
    )
    comment_lines.append(
        "#   serviceTask / scriptTask → use CiviRules or a webhook activity"
    )
    comment_lines.append(
        "#   dataObject / dataStore   → use custom fields on the case"
    )
    comment_lines.append(
        "#   timer events             → use ActionSchedule (scheduled reminders)"
    )
    comment_lines.append(
        "#   message/signal events    → use CiviRules event triggers"
    )

    header = "\n".join(comment_lines)

    # ── Lane → role note ──────────────────────────────────────────────────
    lane_notes = ""
    if proc.lanes:
        lane_notes = "\n# Lanes mapped to caseRoles:\n"
        for lane in proc.lanes:
            tasks_in_lane = [t.name for t in proc.tasks if t.lane == lane.name]
            if tasks_in_lane:
                task_list = ", ".join(f'"{n}"' for n in tasks_in_lane)
                lane_notes += f"#   {lane.name}: {task_list}\n"

    # ── JSON definition ───────────────────────────────────────────────────
    def_json = json.dumps(definition, ensure_ascii=False, indent=2)
    # Indent all lines except the first for HCL alignment
    def_lines = def_json.splitlines()
    def_indented = def_lines[0] + "\n" + textwrap.indent("\n".join(def_lines[1:]), "    ")

    tf = f"""{header}
{lane_notes}
resource "civicrm_case_type" "{resource_name}" {{
  name        = "{_tf_string(resource_name)}"
  title       = "{_tf_string(title)}"
  description = "{_tf_string(description)}"
  is_active   = true
  is_reserved = false
  weight      = {weight}

  definition = jsonencode({def_indented})
}}
"""
    return tf


def render_mapping_report(proc: BpmnProcess, definition: dict) -> str:
    lines = ["=" * 60, "BPMN → CiviCase Mapping Report", "=" * 60, ""]

    lines.append(f"Process: {proc.name} (id={proc.id})")
    lines.append("")

    lines.append(f"Tasks found: {len(proc.tasks)}")
    for t in proc.tasks:
        lane_info = f" [lane: {t.lane}]" if t.lane else ""
        par_info = f" [parallel group: {t.parallel_group}]" if t.parallel_group else ""
        timer_info = f" [timer: {t.timer_offset}d]" if t.timer_offset else ""
        desc_info = f" — {t.description}" if t.description else ""
        lines.append(f"  • [{t.tag}] {t.name}{lane_info}{par_info}{timer_info}{desc_info}")

    lines.append("")
    lines.append(f"Gateways found: {len(proc.gateways)}")
    for gw in proc.gateways.values():
        lines.append(f"  • [{gw.tag}] {gw.id} ({gw.name or 'unnamed'})")

    lines.append("")
    lines.append(f"Lanes (→ caseRoles): {len(proc.lanes)}")
    for lane in proc.lanes:
        lines.append(f"  • {lane.name} ({len(lane.flow_node_refs)} nodes)")

    lines.append("")
    lines.append(f"Generated activityTypes: {len(definition['activityTypes'])}")
    for at in definition["activityTypes"]:
        lines.append(f"  • {at['name']}")

    lines.append("")
    lines.append(f"Timeline entries: {len(definition['activitySets'][0]['activityTypes'])}")
    for entry in definition["activitySets"][0]["activityTypes"]:
        offset = entry.get("reference_offset", "0")
        lines.append(f"  • Day {offset:>3}: {entry['name']}")

    lines.append("")
    lines.append(f"Case Roles: {len(definition['caseRoles'])}")
    for role in definition["caseRoles"]:
        lines.append(f"  • {role['name']}")

    if proc.warnings:
        lines.append("")
        lines.append("Warnings:")
        for w in proc.warnings:
            lines.append(f"  ! {w}")

    lines.append("")
    lines.append("Unmapped BPMN features → CiviCase suggestions:")
    lines.append("  parallelGateway  → same timeline offset; enforce via CiviRules")
    lines.append("  exclusiveGateway → multiple activitySets or CiviRules conditions")
    lines.append("  serviceTask      → webhook activity or CiviRules action")
    lines.append("  dataObject       → custom fields (civicrm_custom_field)")
    lines.append("  timerEvent       → ActionSchedule (civicrm_action_schedule)")
    lines.append("  messageEvent     → CiviRules trigger + action")

    return "\n".join(lines)


# ─── CLI ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(
        description="Convert a BPMN file to a civicrm_case_type Terraform resource.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    parser.add_argument("bpmn_file", help="Path to the BPMN input file")
    parser.add_argument(
        "--output", "-o",
        help="Output .tf file path (default: <bpmn_file_stem>.tf)",
    )
    parser.add_argument(
        "--weight", type=int, default=1,
        help="Sort weight for the case type (default: 1)",
    )
    parser.add_argument(
        "--print-mapping", action="store_true",
        help="Print the BPMN→CiviCase mapping report to stdout",
    )
    parser.add_argument(
        "--stdout", action="store_true",
        help="Print Terraform HCL to stdout instead of writing a file",
    )
    args = parser.parse_args()

    bpmn_path = Path(args.bpmn_file)
    if not bpmn_path.exists():
        sys.exit(f"ERROR: File not found: {bpmn_path}")

    proc = parse_bpmn(bpmn_path)
    definition = build_civicase_definition(proc)

    if args.print_mapping:
        print(render_mapping_report(proc, definition))
        return

    tf_content = render_terraform(proc, definition, args.weight)

    if args.stdout:
        print(tf_content)
    else:
        out_path = Path(args.output) if args.output else bpmn_path.with_suffix(".tf")
        out_path.write_text(tf_content, encoding="utf-8")
        print(f"Written: {out_path}")

    if proc.warnings:
        print("\nWarnings:", file=sys.stderr)
        for w in proc.warnings:
            print(f"  ! {w}", file=sys.stderr)


if __name__ == "__main__":
    main()
