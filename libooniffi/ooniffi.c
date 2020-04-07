#include "ooniffi.h"

#include <stdlib.h>

#include "_cgo_export.h"

struct mk_events_{};
struct mk_task_{};

ooniffi_task_t *ooniffi_task_start(const char *settings) {
    /* Implementation note: Go does not have the concept of const but
       we know that the code is just making a copy of settings. */
    return (ooniffi_task_t *)ooniffi_cgo_task_start((char *)settings);
}

ooniffi_event_t *ooniffi_task_wait_for_next_event(ooniffi_task_t *task) {
    return (ooniffi_event_t *)ooniffi_cgo_task_wait_for_next_event((intptr_t)task);
}

int ooniffi_task_is_done(ooniffi_task_t *task) {
    return ooniffi_cgo_task_is_done((intptr_t)task);
}

void ooniffi_task_interrupt(ooniffi_task_t *task) {
    ooniffi_cgo_task_interrupt((intptr_t)task);
}

const char *ooniffi_event_serialization(ooniffi_event_t *event) {
    return (const char *)event;
}

void ooniffi_event_destroy(ooniffi_event_t *event) {
    ooniffi_cgo_event_destroy((char *)event);
}

void ooniffi_task_destroy(ooniffi_task_t *task) {
    ooniffi_cgo_task_destroy((intptr_t)task);
}
