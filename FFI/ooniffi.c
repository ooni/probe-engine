#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>

#include "ooniffi.h"

static void errx(int exitcode, const char *format, ...) {
  va_list ap;
  va_start(ap, format);
  (void)vfprintf(stderr, format, ap);
  va_end(ap);
  exit(exitcode);
}

int main(int argc, const char *const *argv) {
  if (argc != 2) {
    errx(1, "usage: %s <config-file>\n", argv[0]);
  }
  FILE *filep = fopen(argv[1], "rb");
  if (filep == NULL) {
    errx(1, "cannot open: %s", argv[1]);
  }
  const size_t bufsiz = 1 << 1;
  char *settings = calloc(1, bufsiz);
  if (settings == NULL) {
    errx(1, "cannot allocate memory\n");
  }
  size_t n = fread(settings, 1, bufsiz, filep);
  if (n >= bufsiz || !feof(filep)) {
    errx(1, "cannot read file until EOF\n");
  }
  settings[n] = '\0';
  ooniffi_task_t *task = ooniffi_task_start(settings);
  while (!ooniffi_task_is_done(task)) {
    ooniffi_event_t *event = ooniffi_task_wait_for_next_event(task);
    printf("%s\n", ooniffi_event_serialization(event));
    ooniffi_event_destroy(event);
  }
  ooniffi_task_destroy(task);
  exit(0);
}
