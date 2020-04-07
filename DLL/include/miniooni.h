#ifndef INCLUDE_MINIOONI_H_
#define INCLUDE_MINIOONI_H_

#include <stdint.h>
#include <stdlib.h>

typedef struct miniooni_task_ miniooni_task_t;
typedef struct miniooni_event_ miniooni_event_t;

#ifdef __cplusplus
extern "C" {
#endif

extern miniooni_task_t *miniooni_task_start(const char *settings);
extern miniooni_event_t *miniooni_task_wait_for_next_event(miniooni_task_t *task);
extern int miniooni_task_is_done(miniooni_task_t *task);
extern void miniooni_task_interrupt(miniooni_task_t *task);
extern const char *miniooni_event_serialization(miniooni_event_t *event);
extern void miniooni_event_destroy(miniooni_event_t *event);
extern void miniooni_task_destroy(miniooni_task_t *task);

#ifdef __cplusplus
}
#endif
#endif /* INCLUDE_MINIOONI_H_ */
