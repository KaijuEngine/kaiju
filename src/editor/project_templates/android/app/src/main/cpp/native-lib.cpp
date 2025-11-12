#include <android_native_app_glue.h>
#include "libkaiju_android.h"

void android_main(struct android_app* state) {
    app_dummy();
    AndroidMain(state);
}