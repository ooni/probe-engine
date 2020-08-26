#ifndef INCLUDE_OONIFFI_H_
#define INCLUDE_OONIFFI_H_

#include <stdint.h>
#include <stdlib.h>

/*
 * API/ABI compatible with Measurement Kit v0.10.11 [1].
 *
 * Just replace `mk_` with `ooniffi_` and recompile.
 *
 * .. [1] https://github.com/measurement-kit/measurement-kit/tree/v0.10.11/
 */

typedef struct ooniffi_task_ ooniffi_task_t;
typedef struct ooniffi_event_ ooniffi_event_t;

#ifdef __cplusplus
extern "C" {
#endif

extern ooniffi_task_t *ooniffi_task_start(const char *settings);
extern ooniffi_event_t *ooniffi_task_wait_for_next_event(ooniffi_task_t *task);
extern int ooniffi_task_is_done(ooniffi_task_t *task);
extern void ooniffi_task_interrupt(ooniffi_task_t *task);
extern const char *ooniffi_event_serialization(ooniffi_event_t *str);
extern void ooniffi_event_destroy(ooniffi_event_t *str);
extern void ooniffi_task_destroy(ooniffi_task_t *task);

#ifdef __cplusplus
}
#endif

/* Define OONIFFI_EMULATE_MK_API to provide MK compatible API. */
#ifdef OONIFFI_EMULATE_MK_API
#define mk_task_start ooniffi_task_start
#define mk_task_wait_for_next_event ooniffi_wait_for_next_event
#define mk_task_is_done ooniffi_task_is_done
#define mk_task_interrupt ooniffi_task_interrupt
#define mk_event_serialization ooniffi_event_serialization
#define mk_event_destroy ooniffi_event_destroy
#define mk_task_destroy ooniffi_task_destroy
#endif

#endif /* INCLUDE_OONIFFI_H_ */
