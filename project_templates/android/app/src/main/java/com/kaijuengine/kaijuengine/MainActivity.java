package com.kaijuengine.kaijuengine;

import android.app.NativeActivity;

import android.os.Bundle;
import android.view.View;

public class MainActivity extends NativeActivity {
    static {
        System.loadLibrary("kaijuengine");
        System.loadLibrary("kaiju_android");
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        int SDK_INT = android.os.Build.VERSION.SDK_INT;
        View decorView = getWindow().getDecorView();
        super.onCreate(savedInstanceState);
    }
}