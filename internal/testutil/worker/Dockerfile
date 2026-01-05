# Test Worker Image for Bosun Integration Tests
#
# This image provides a minimal worker that validates Bosun's job execution:
# - Prints environment variables set by Bosun
# - Lists mounted volumes to verify attachment
# - Exits successfully (code 0)
#
# Build:
#   docker build -t backup-worker:test internal/testutil/worker/
#
# TODO: Document test worker in README or docs/testing.md
# TODO: Add failure mode variants (exit 1, timeout, etc.) for error path testing
# TODO: Consider publishing to a test registry for CI reproducibility

FROM alpine:latest

RUN apk add --no-cache bash

ENTRYPOINT ["/bin/sh", "-c"]

CMD ["echo '=== BOSUN WORKER TEST ===' && \
     echo '' && \
     echo '--- Environment ---' && \
     echo \"BOSUN_JOB_NAME:  $BOSUN_JOB_NAME\" && \
     echo \"BOSUN_RUN_ID:    $BOSUN_RUN_ID\" && \
     echo \"BOSUN_STACK:     $BOSUN_STACK\" && \
     echo \"BOSUN_DRY_RUN:   $BOSUN_DRY_RUN\" && \
     echo '' && \
     echo '--- Mounted Volumes ---' && \
     if [ -d /backup ]; then ls -la /backup/; else echo 'No /backup directory'; fi && \
     echo '' && \
     echo '=== WORKER SUCCESS ===' && \
     exit 0"]
