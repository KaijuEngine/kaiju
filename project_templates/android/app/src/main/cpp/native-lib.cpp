#include <jni.h>
#include <string>
#include "libkaiju_android.h"

extern "C" JNIEXPORT jstring JNICALL
Java_com_kaijuengine_kaijuengine_MainActivity_stringFromJNI(JNIEnv* env, jobject /* this */) {
	AndroidMain(nullptr);
    std::string hello = "Hello from C++";
    return env->NewStringUTF(hello.c_str());
}