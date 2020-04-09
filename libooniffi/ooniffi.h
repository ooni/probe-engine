#ifndef INCLUDE_OONIFFI_H_
#define INCLUDE_OONIFFI_H_

#include <stdint.h>
#include <stdlib.h>

/**
 * @file ooniffi.h
 * @brief github.com/ooni/probe-engine FFI friendly API.
 * Usage is as follows:
 *
 * ```C
 * intptr_t task = ooniffi_task_start(settings);
 * if (task == 0) {
 *     return;
 * }
 * while (!ooniffi_task_done(task)) {
 *     char *ev = ooniffi_task_yield_from(task);
 *     if (ev != NULL) {
 *         printf("%s\n", ev);
 *     }
 *     ooniffi_string_free(ev);
 * }
 * ooniffi_task_destroy(task);
 * ```
 *
 * Where settings and ev are serialized JSONs following [the
 * specification of Measurement Kit v0.10.11](
 * https://github.com/measurement-kit/measurement-kit/tree/v0.10.11/include/measurement_kit).
 */

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief Starts a OONI measurement task.
 * @param settings serialized JSON containing settings compatible with [the
 * specification of Measurement Kit v0.10.11](
 * https://github.com/measurement-kit/measurement-kit/tree/v0.10.11/include/measurement_kit).
 * @return nonzero on success.
 * @return zero if @p settings is NULL or if @p settings does not contain a
 * valid serialized JSON.
 */
extern intptr_t ooniffi_task_start(const char *settings);

/**
 * @brief Waits for the next event emitted by @p task.
 * @return NULL if the task is zero.
 * @return a valid serialized JSON event compatible with
 * [the specification of Measurement Kit v0.10.11](
 * https://github.com/measurement-kit/measurement-kit/tree/v0.10.11/include/measurement_kit)
 * otherwise.
 * @remark You must free the returned string with ooniffi_string_free.
 */
extern char *ooniffi_task_yield_from(intptr_t task);

/**
 * @brief Tells you whether @p task is done.
 * @return nonzero if @p task is done or @p task is zero.
 * @return zero otherwise.
 */
extern int ooniffi_task_done(intptr_t task);

/**
 * @brief Tells @p task to stop as soon as possible.
 * @remark This function does nothing if @p task is zero.
 */
extern void ooniffi_task_interrupt(intptr_t task);

/** @brief Frees the provided @p str.
 * @remark This function does nothing if @p str is NULL.
 */
extern void ooniffi_string_free(char *str);

/**
 * @brief Interrupts @p task and releases the resources used by it.
 * @remark This function does nothing if @p task is zero.
 */
extern void ooniffi_task_destroy(intptr_t task);

#ifdef __cplusplus
}
#endif
#endif /* INCLUDE_OONIFFI_H_ */
