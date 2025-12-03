plugins {
    alias(libs.plugins.android.application)
}

android {
    namespace = "com.kaijuengine.kaijuengine"
    compileSdk = 34

    sourceSets {
        named("main") {
            jniLibs.srcDir("src/main/jniLibs")
        }
    }

    defaultConfig {
        applicationId = "com.kaijuengine.kaijuengine"
        minSdk = 28
        targetSdk = 34
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        ndk {
            abiFilters += setOf("arm64-v8a")
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
            packaging {
                jniLibs.excludes.add("**/libVkLayer_khronos_validation.so")
            }
        }
        debug {
            packaging {
                jniLibs.excludes.remove("**/libVkLayer_khronos_validation.so")
            }
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }
    externalNativeBuild {
        cmake {
            path = file("src/main/cpp/CMakeLists.txt")
            version = "3.22.1"
        }
    }
    buildFeatures {
        viewBinding = true
    }
}

dependencies {

    implementation(libs.appcompat)
    implementation(libs.material)
    implementation(libs.constraintlayout)
    testImplementation(libs.junit)
    androidTestImplementation(libs.ext.junit)
    androidTestImplementation(libs.espresso.core)
}