#include "ffi.h"

#include <limits.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#include "_cgo_export.h"

struct ooni_task_ {
	intptr_t handle;
};

struct ooni_event_ {
	char   *base;
	size_t length;
};

ooni_task_t *ooni_task_start(const char *settings) {
	ooni_task_t *tap = calloc(1, sizeof(*tap));
	if (tap == NULL) {
		return NULL;
	}
	if ((tap->handle = OONIGoTaskStart((char *)settings)) == 0) {
		free(tap);
		return NULL;
	}
	return tap;
}

ooni_event_t *ooni_task_wait_for_next_event(ooni_task_t *tap) {
	if (tap == NULL) {
		return NULL;
	}
	ooni_event_t *evp = calloc(1, sizeof(*evp));
	if (evp == NULL) {
		return NULL;
	}
	if (OONIGoTaskWaitForNextEvent(
				tap->handle, &evp->base, &evp->length) == 0) {
		free(evp);
		return NULL;
	}
	return evp;
}

int ooni_task_is_done(ooni_task_t *tap) {
	return (tap != NULL) ? OONIGoTaskIsDone(tap->handle) : 1;
}

void ooni_task_interrupt(ooni_task_t *tap) {
	if (tap != NULL) {
		OONIGoTaskInterrupt(tap->handle);
	}
}

const char *ooni_event_serialization(ooni_event_t *evp) {
	return (evp != NULL) ? evp->base : NULL;
}

size_t ooni_event_serialization_size(ooni_event_t *evp) {
	return (evp != NULL) ? evp->length : 0;
}

void ooni_event_destroy(ooni_event_t *evp) {
	if (evp != NULL) {
		free(evp->base);
		free(evp);
	}
}

void ooni_task_destroy(ooni_task_t *tap) {
	if (tap != NULL) {
		OONIGoTaskDestroy(tap->handle);
		free(tap);
	}
}
