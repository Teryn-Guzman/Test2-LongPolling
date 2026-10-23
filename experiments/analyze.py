#!/usr/bin/env python3
"""Summarize actual ImageLab Assessment 2 client/server JSONL logs."""

import argparse
import csv
import json
import statistics
from collections import defaultdict
from datetime import datetime
from pathlib import Path

CSV_COLUMNS = [
    "run_id", "condition", "mode", "repetition", "job_id", "image",
    "latency_setting", "baseline_version", "terminal_version",
    "requests_started", "initial_200", "newer_200", "unchanged_200",
    "no_change_204", "retrieval_errors", "cancellations",
    "distinct_new_versions_seen", "processing_ms", "detection_ms", "notes",
    "observation_db_reads",
]
CONDITIONS = ["A", "B", "C1", "C2", "C3"]
MODES = ["short", "long"]


def read_jsonl(path):
    rows = []
    with Path(path).open(encoding="utf-8") as stream:
        for line_number, line in enumerate(stream, 1):
            if not line.strip():
                continue
            try:
                rows.append(json.loads(line))
            except json.JSONDecodeError as error:
                raise SystemExit(f"Invalid JSON in {path}:{line_number}: {error}")
    return rows


def epoch_ms(value):
    if value is None or value == "":
        return None
    if isinstance(value, (int, float)):
        return float(value)
    parsed = datetime.fromisoformat(str(value).replace("Z", "+00:00"))
    return parsed.timestamp() * 1000


def summarize(client_events, server_events):
    runs = defaultdict(list)
    for event in client_events:
        if event.get("run_id"):
            runs[event["run_id"]].append(event)

    server_by_job = defaultdict(list)
    for event in server_events:
        job_id = event.get("job_id")
        if job_id:
            server_by_job[job_id].append(event)

    output = []
    for run_id, events in sorted(runs.items()):
        events.sort(key=lambda item: item.get("client_time_ms", 0))
        metadata = next((event for event in events if event.get("job_id")), events[0])
        job_id = metadata.get("job_id", "")
        request_events = [event for event in events if event.get("event") == "request_start"]
        responses = [event for event in events if event.get("event") == "response_received"]
        initial = sum(event.get("classification") == "initial_200" for event in responses)
        newer = sum(event.get("classification") == "newer_200" for event in responses)
        unchanged = sum(event.get("classification") == "unchanged_200" for event in responses)
        no_change = sum(event.get("http_status") == 204 for event in responses)
        errors = [event for event in events if event.get("event") == "retrieval_error"]
        cancellations = [event for event in events if event.get("event") == "request_cancelled"]
        received_versions = {
            event.get("received_version") for event in responses
            if event.get("classification") == "newer_200" and event.get("received_version") is not None
        }
        committed_versions = {
            event.get("version") for event in server_by_job.get(job_id, [])
            if event.get("msg") == "job committed" and isinstance(event.get("version"), int)
        }
        baseline = next((event.get("received_version") for event in responses if event.get("classification") == "initial_200"), "")
        terminal = next((event for event in reversed(responses) if event.get("job_status") in ("completed", "failed")), None)
        terminal_version = terminal.get("received_version", "") if terminal else ""
        terminal_status = terminal.get("job_status", "") if terminal else ""

        processing_ms = ""
        detection_ms = ""
        if terminal:
            terminal_time = terminal.get("completed_at") or terminal.get("failed_at")
            if terminal.get("started_at") and terminal_time:
                processing_ms = round(epoch_ms(terminal_time) - epoch_ms(terminal["started_at"]))

            candidates = [
                event for event in server_by_job.get(job_id, [])
                if event.get("msg") == "job committed"
                and event.get("status") == terminal_status
                and event.get("version") == terminal_version
                and event.get("post_commit_at")
            ]
            if candidates:
                terminal_receipt = terminal.get("client_time_ms")
                if terminal_receipt is not None:
                    detection_ms = round(terminal_receipt - epoch_ms(candidates[-1]["post_commit_at"]))

        completed_request_ids = {
            event.get("request_id") for event in events
            if event.get("event") in ("response_received", "retrieval_error", "request_cancelled")
        }
        pending = len({event.get("request_id") for event in request_events} - completed_request_ids)
        reads = sum(
            int(event.get("queries", 0)) for event in server_by_job.get(job_id, [])
            if event.get("msg") == "observation database reads"
        )
        job_times = [epoch_ms(event.get("time")) for event in server_by_job.get(job_id, []) if event.get("time")]
        first_job_time = min(job_times) if job_times else None
        settings = [event for event in server_events if event.get("msg") == "experiment workload configuration"]
        eligible_settings = [event for event in settings if first_job_time is not None and epoch_ms(event.get("time")) <= first_job_time]
        workload = max(eligible_settings, key=lambda event: epoch_ms(event.get("time"))) if eligible_settings else {}
        notes = []
        if pending:
            notes.append(f"{pending} request(s) still pending when log ended")
        if terminal and detection_ms == "":
            notes.append("terminal server post-commit log not joined")
        if detection_ms != "" and detection_ms < 0:
            notes.append("negative delay; investigate wall-clock/log ordering")
        if received_versions:
            notes.append("new versions received=" + ",".join(map(str, sorted(received_versions))))
        if committed_versions:
            notes.append("job versions committed=" + ",".join(map(str, sorted(committed_versions))))
        if baseline != "" and terminal_version != "":
            committed_observable = {version for version in committed_versions if baseline < version <= terminal_version}
            gaps = sorted(committed_observable - received_versions)
            notes.append("intermediate versions not received=" + (",".join(map(str, gaps)) if gaps else "none"))
        if workload:
            notes.append(
                "server workload: quiet_delay={quiet}, simulation_duration={duration}, simulation_interval={interval}".format(
                    quiet=workload.get("quiet_delay", "unknown"),
                    duration=workload.get("simulation_duration", "unknown"),
                    interval=workload.get("simulation_interval", "unknown"),
                )
            )
        output.append({
            "run_id": run_id,
            "condition": metadata.get("condition", ""),
            "mode": metadata.get("mode", ""),
            "repetition": metadata.get("repetition", ""),
            "job_id": job_id,
            "image": metadata.get("image", ""),
            "latency_setting": metadata.get("latency_setting", ""),
            "baseline_version": baseline,
            "terminal_version": terminal_version,
            "requests_started": len(request_events),
            "initial_200": initial,
            "newer_200": newer,
            "unchanged_200": unchanged,
            "no_change_204": no_change,
            "retrieval_errors": len(errors),
            "cancellations": len(cancellations),
            "distinct_new_versions_seen": len(received_versions),
            "processing_ms": processing_ms,
            "detection_ms": detection_ms,
            "notes": "; ".join(notes),
            "observation_db_reads": reads,
        })
    return output


def median_range(values):
    values = [float(value) for value in values if value != ""]
    if not values:
        return "n/a"
    median = statistics.median(values)
    return f"{median:g} ({min(values):g}–{max(values):g})"


def write_svg(path, title, metric, unit, rows):
    width, height = 920, 470
    left, right, top, bottom = 180, 34, 62, 48
    chart_width = width - left - right
    chart_height = height - top - bottom
    groups = [(condition, mode, median_range([row[metric] for row in rows if row["condition"] == condition and row["mode"] == mode]))
              for condition in CONDITIONS for mode in MODES]
    numeric = [float(row[metric]) for row in rows if row[metric] != ""]
    maximum = max(numeric, default=1) or 1
    bar_height = 19
    gap = 8
    group_gap = 15
    current_y = top
    elements = [f'<text x="{width/2}" y="28" text-anchor="middle" class="title">{title}</text>']
    for condition in CONDITIONS:
        elements.append(f'<text x="{left-14}" y="{current_y+27}" text-anchor="end" class="condition">{condition}</text>')
        for mode in MODES:
            vals = [float(row[metric]) for row in rows if row["condition"] == condition and row["mode"] == mode and row[metric] != ""]
            value = statistics.median(vals) if vals else 0
            bar_width = chart_width * value / maximum
            color = "#5a63e0" if mode == "short" else "#27a9bd"
            label = "n/a" if not vals else f"{value:g} {unit}"
            elements.append(f'<rect x="{left}" y="{current_y}" width="{bar_width:.1f}" height="{bar_height}" rx="4" fill="{color}"/>')
            elements.append(f'<text x="{left+bar_width+7:.1f}" y="{current_y+15}" class="value">{label}</text>')
            elements.append(f'<text x="{left-8}" y="{current_y+14}" text-anchor="end" class="mode">{mode}</text>')
            current_y += bar_height + gap
        current_y += group_gap
    legend_y = height - 14
    elements.extend([
        f'<rect x="{left}" y="{legend_y-12}" width="12" height="12" fill="#5a63e0"/><text x="{left+18}" y="{legend_y-2}" class="legend">Short polling</text>',
        f'<rect x="{left+130}" y="{legend_y-12}" width="12" height="12" fill="#27a9bd"/><text x="{left+148}" y="{legend_y-2}" class="legend">Long polling</text>',
        f'<text x="{width-right}" y="{height-2}" text-anchor="end" class="axis">Median per condition/mode; bars use individual-run median</text>',
    ])
    svg = f'''<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">
<style>text{{font-family:Arial,sans-serif;fill:#263244}}.title{{font-size:19px;font-weight:bold}}.condition{{font-size:14px;font-weight:bold}}.mode,.value,.legend,.axis{{font-size:11px}}.value{{fill:#111827}}</style>
<rect width="100%" height="100%" fill="white"/>{''.join(elements)}</svg>'''
    Path(path).write_text(svg, encoding="utf-8")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--client", required=True, help="Downloaded client JSONL observations")
    parser.add_argument("--server", required=True, help="Server JSONL stdout log")
    parser.add_argument("--out-dir", default="experiments/results", help="Output directory")
    args = parser.parse_args()
    rows = summarize(read_jsonl(args.client), read_jsonl(args.server))
    if not rows:
        raise SystemExit("No client run records found; no results were generated.")
    output_dir = Path(args.out_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    with (output_dir / "summary.csv").open("w", newline="", encoding="utf-8") as stream:
        writer = csv.DictWriter(stream, fieldnames=CSV_COLUMNS)
        writer.writeheader()
        writer.writerows(rows)

    lines = ["# Assessment 2 measured comparison", "", "Every row below is generated from the supplied client/server logs. Blank detection values mean the terminal commit could not be joined. The version lists and gaps are retained in the notes column.", "", "## Runs", "", "| Run | Condition | Mode | Rep | Job | Requests | Detection (ms) | Versions | Notes |", "|---|---|---|---:|---|---:|---:|---:|---|"]
    for row in rows:
        lines.append("| {run_id} | {condition} | {mode} | {repetition} | {job_id} | {requests_started} | {detection_ms} | {distinct_new_versions_seen} | {notes} |".format(**row))
    lines.extend(["", "## Median (minimum–maximum) by condition and mode", "", "| Condition | Mode | Requests | Detection delay (ms) |", "|---|---|---:|---:|"])
    for condition in CONDITIONS:
        for mode in MODES:
            selected = [row for row in rows if row["condition"] == condition and row["mode"] == mode]
            lines.append(f"| {condition} | {mode} | {median_range([row['requests_started'] for row in selected])} | {median_range([row['detection_ms'] for row in selected])} |")
    lines.extend(["", "## Run-count check", "", f"Observed {len(rows)} runs; required minimum is 30 (3 repetitions × 5 conditions × 2 modes)."])
    for condition in CONDITIONS:
        for mode in MODES:
            count = sum(row["condition"] == condition and row["mode"] == mode for row in rows)
            if count < 3:
                lines.append(f"- {condition}/{mode}: {count} run(s), fewer than the required 3.")
    (output_dir / "summary.md").write_text("\n".join(lines) + "\n", encoding="utf-8")
    write_svg(output_dir / "request-count-by-condition.svg", "Observation request count by condition", "requests_started", "requests", rows)
    write_svg(output_dir / "detection-delay-by-condition.svg", "Completion-detection delay by condition", "detection_ms", "ms", rows)
    print(f"Wrote {len(rows)} measured run rows to {output_dir}; inspect join notes and run-count warnings before submission.")


if __name__ == "__main__":
    main()
