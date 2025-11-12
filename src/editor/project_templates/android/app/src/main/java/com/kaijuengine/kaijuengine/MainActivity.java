package com.kaijuengine.kaijuengine;

import android.app.NativeActivity;

import android.content.Intent;
import android.net.Uri;
import android.os.Bundle;
import android.util.DisplayMetrics;
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

	public void open_web_url(String url) {
        Uri webpage = Uri.parse(url);
        Intent intent = new Intent(Intent.ACTION_VIEW, webpage);
        if (intent.resolveActivity(getPackageManager()) != null) {
            startActivity(intent);
        }
    }

	public float width_mm() {
		DisplayMetrics dm = getResources().getDisplayMetrics();
		return (dm.widthPixels / dm.xdpi) * 25.4F;
	}

	public float height_mm() {
		DisplayMetrics dm = getResources().getDisplayMetrics();
		return (dm.heightPixels / dm.ydpi) * 25.4F;
	}
}