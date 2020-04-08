#include "ooniffi.h"

#include <stdlib.h>

#include "_cgo_export.h"

struct mk_events_{
    intptr_t handle;
};

struct mk_task_{
    char *string;
};

ooniffi_task_t *ooniffi_task_start(const char *settings) {
    if (settings == NULL) {
        return NULL;
    }
    ooniffi_task_t *task = calloc(1, sizeof(*task));
    if (task == NULL) {
        return NULL;
    }
    /* Implementation note: Go does not have the concept of const but
       we know that the code is just making a copy of settings. */
    task->handle = ooniffi_cgo_task_start((char *)settings);
    if (task->handle == 0) {
        free(task);
        return NULL;
    }
    return task;
}

ooniffi_event_t *ooniffi_task_wait_for_next_event(ooniffi_task_t *task) {
    if (task == NULL) {
        return NULL;
    }
    ooniffi_event_t *event = calloc(1, sizeof(*event));
    if (event == NULL) {
        return NULL;
    }
    /* "As a special case, C.malloc does not call the C library malloc directly but
       instead calls a Go helper function that wraps the C library malloc but guarantees
       never to return nil. If C's malloc indicates out of memory, the helper function
       crashes the program, like when Go itself runs out of memory." */
    event->string = ooniffi_cgo_task_wait_for_next_event(task->handle);
    return event;
}

int ooniffi_task_is_done(ooniffi_task_t *task) {
    return ooniffi_cgo_task_is_done((task != NULL) ? task->handle : 0);
}

void ooniffi_task_interrupt(ooniffi_task_t *task) {
    ooniffi_cgo_task_interrupt((task != NULL) ? task->handle : 0);
}

const char *ooniffi_event_serialization(ooniffi_event_t *event) {
    return (event != NULL) ? event->string : NULL;
}

void ooniffi_event_destroy(ooniffi_event_t *event) {
    /* We assume that free handles NULL, which is now the case everywhere */
    ooniffi_cgo_event_destroy((event != NULL) ? event->string : NULL);
    free(event);
}

void ooniffi_task_destroy(ooniffi_task_t *task) {
    ooniffi_cgo_task_destroy((task != NULL) ? task->handle : 0);
    free(task);
}
