#include "DLL/include/miniooni.h"

#include <stdlib.h>

#include "_cgo_export.h"

struct mk_events_{};
struct mk_task_{};

miniooni_task_t *miniooni_task_start(const char *settings) {
    /* Implementation note: Go does not have the concept of const but
       we know that the code is just making a copy of settings. */
    return (miniooni_task_t *)miniooni_cgo_task_start((char *)settings);
}

miniooni_event_t *miniooni_task_wait_for_next_event(miniooni_task_t *task) {
    return (miniooni_event_t *)miniooni_cgo_task_wait_for_next_event((intptr_t)task);
}

int miniooni_task_is_done(miniooni_task_t *task) {
    return miniooni_cgo_task_is_done((intptr_t)task);
}

void miniooni_task_interrupt(miniooni_task_t *task) {
    miniooni_cgo_task_interrupt((intptr_t)task);
}

const char *miniooni_event_serialization(miniooni_event_t *event) {
    return (const char *)event;
}

void miniooni_event_destroy(miniooni_event_t *event) {
    miniooni_cgo_event_destroy((char *)event);
}

void miniooni_task_destroy(miniooni_task_t *task) {
    miniooni_cgo_task_destroy((intptr_t)task);
}