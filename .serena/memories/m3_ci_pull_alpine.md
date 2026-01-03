Decision: Add CI step to pre-pull the Alpine worker image

Reasoning:
- Integration tests assume the worker image (alpine:latest) is available locally. In CI (GH Actions) the image may not be cached which causes M3 tests to fail due to executor's image validation (ImageInspect only).
- The project follows BYOI (Bring Your Own Image) philosophy; for typical users, worker images are built/pulled by the user and available locally.

Action taken:
- Added `docker pull alpine:latest` step to `.github/workflows/ci.yml` before running integration tests to ensure the placeholder worker image is available for test runs.

Follow-up:
- Created issue to track implementing optional image pull support in the executor for M3 as a nice-to-have; this is intentionally not implemented by default to respect BYOI and avoid changing runtime semantics without design.

Date: 2026-01-03
