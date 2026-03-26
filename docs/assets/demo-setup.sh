#!/bin/bash
# Helper for VHS demo — stores structured migration context
stateprim set migration context '{"blocked_on":"fk_ordering","tables_done":["users","sessions","tokens","events"],"notes":["check constraint order on accounts table","fk to payments must come after billing","do not drop old columns until step 7"]}'
