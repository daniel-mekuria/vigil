"""Standard-library checks for the copyable StatLite Metrics helper."""

from __future__ import annotations

import asyncio
import sys
import unittest
from pathlib import Path


sys.path.insert(0, str(Path(__file__).parent))

from statlite_metrics import DEFAULT_METRICS_PATH, SCHEMA, StatLiteMetrics


class _Request:
    def __init__(self, path: str) -> None:
        self.url = type("URL", (), {"path": path})()


class _Response:
    status_code = 200


class StatLiteMetricsTests(unittest.TestCase):
    def test_snapshot_is_application_and_process_profile_without_host_fields(self) -> None:
        snapshot = StatLiteMetrics().snapshot()

        self.assertEqual(snapshot["schema"], SCHEMA)
        self.assertEqual(snapshot["status"], "UP")
        self.assertIn("started_at", snapshot)
        metrics = snapshot["metrics"]
        self.assertIn("process_cpu_usage", metrics)
        self.assertNotIn("cpu_usage", metrics)
        self.assertIn("runtime_heap_used_bytes", metrics)
        for key in (
            "host_cpu_usage",
            "host_memory_used_bytes",
            "host_memory_total_bytes",
            "host_disk_used_bytes",
            "host_disk_total_bytes",
        ):
            self.assertNotIn(key, metrics)

    def test_metrics_scrape_does_not_increment_application_counters(self) -> None:
        helper = StatLiteMetrics()

        async def call_next(_request: object) -> _Response:
            return _Response()

        asyncio.run(helper.middleware(_Request(DEFAULT_METRICS_PATH), call_next))
        self.assertEqual(helper.snapshot()["metrics"]["requests_total"], 0)

        asyncio.run(helper.middleware(_Request("/"), call_next))
        self.assertEqual(helper.snapshot()["metrics"]["requests_total"], 1)
