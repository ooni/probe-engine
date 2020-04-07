#ifndef INCLUDE_OONIFFI_H_
#define INCLUDE_OONIFFI_H_

#include <stdint.h>
#include <stdlib.h>

typedef struct ooniffi_task_ ooniffi_task_t;
typedef struct ooniffi_event_ ooniffi_event_t;

#ifdef __cplusplus
extern "C" {
#endif

extern ooniffi_task_t *ooniffi_task_start(const char *settings);
extern ooniffi_event_t *ooniffi_task_wait_for_next_event(ooniffi_task_t *task);
extern int ooniffi_task_is_done(ooniffi_task_t *task);
extern void ooniffi_task_interrupt(ooniffi_task_t *task);
extern const char *ooniffi_event_serialization(ooniffi_event_t *event);
extern void ooniffi_event_destroy(ooniffi_event_t *event);
extern void ooniffi_task_destroy(ooniffi_task_t *task);

#ifdef __cplusplus
}
#endif
#endif /* INCLUDE_OONIFFI_H_ */
