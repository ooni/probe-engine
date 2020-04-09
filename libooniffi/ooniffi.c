#include "ooniffi.h"

#include "_cgo_export.h"

intptr_t ooniffi_task_start(const char *settings) {
    /* Implementation note: Go does not have the concept of const but
       we know that the code is just making a copy of settings. */
    return ooniffi_task_start_((char *)settings);
}
