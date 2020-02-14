#ifndef LIBOONI_FFI_H
#define LIBOONI_FFI_H

/*-
 * libooni/ffi.h - foreign function interface API for OONI.
 *
 * This API/ABI is compatible with Measurement Kit v0.10.x except that
 * we're using the `ooni_` prefix as opposed to the `mk_` prefix.
 *
 * Also, this API allows you to construct strings more flexibly because
 * you can obtain both an event's string and its size.
 */

#include <stdint.h>
#include <stdlib.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct ooni_task_ ooni_task_t;

typedef struct ooni_event_ ooni_event_t;

extern ooni_task_t *ooni_task_start(const char *settings);

extern ooni_event_t *ooni_task_wait_for_next_event(ooni_task_t *task);

extern int ooni_task_is_done(ooni_task_t *task);

extern void ooni_task_interrupt(ooni_task_t *task);

extern const char *ooni_event_serialization(ooni_event_t *event);

extern size_t ooni_event_serialization_size(ooni_event_t *event);

extern void ooni_event_destroy(ooni_event_t *event);

extern void ooni_task_destroy(ooni_task_t *task);

#ifdef __cplusplus
}
#endif
#endif /* LIBOONI_FFI_H */
