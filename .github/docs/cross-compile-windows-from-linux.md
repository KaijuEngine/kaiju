# 💻 Cross-Compilation Windows from Linux (TODO)

## 🎯 Objective

Eliminate the **Windows runner** from the CI (Continuous Integration) pipeline by performing cross-compilation for Windows from Linux using **MinGW-w64**.

## ✅ Advantages

  * **Build everything on the same Linux runner** (faster)
  * **Less complexity** (no need to manage 2 different OSes)
  * **Cheaper** (Windows runners cost 2x more in CI minutes)
  * **More reliable** (Linux is more stable for CI)

## 🛠️ Recommended Approach

### 1\. Install MinGW on Ubuntu

```bash
sudo apt-get update
sudo apt-get install -y \
  mingw-w64 \
  g++-mingw-w64-x86-64 \
  gcc-mingw-w64-x86-64
```

### 2\. Get the Vulkan Headers

**Option A - Clone from GitHub (recommended)**:

```bash
# Vulkan headers are platform-agnostic
git clone --depth 1 --branch sdk-1.3.290 \
  https://github.com/KhronosGroup/Vulkan-Headers.git
cd Vulkan-Headers
mkdir build && cd build

# Install into the MinGW prefix
cmake -DCMAKE_INSTALL_PREFIX=/usr/x86_64-w64-mingw32 ..
sudo make install
```

**Option B - Copy system headers**:

```bash
# Linux Vulkan headers also work for Windows
sudo apt-get install libvulkan-dev
sudo cp -r /usr/include/vulkan /usr/x86_64-w64-mingw32/include/
sudo cp -r /usr/include/vk_video /usr/x86_64-w64-mingw32/include/
```

### 3\. Get the Windows Vulkan Library

**Option A - Download Windows SDK**:

```bash
wget https://sdk.lunarg.com/sdk/download/1.3.290.0/windows/VulkanSDK-1.3.290.0-Installer.exe
7z x VulkanSDK-1.3.290.0-Installer.exe
sudo cp Lib/vulkan-1.lib /usr/x86_64-w64-mingw32/lib/
```

**Option B - Extract from MSYS2** (more complex):

```bash
# Download the MSYS2 package
wget https://repo.msys2.org/mingw/mingw64/mingw-w64-x86_64-vulkan-loader-*.pkg.tar.zst
tar -I zstd -xf mingw-w64-x86_64-vulkan-loader-*.pkg.tar.zst
sudo cp mingw64/lib/libvulkan.a /usr/x86_64-w64-mingw32/lib/
sudo cp mingw64/bin/vulkan-1.dll /usr/x86_64-w64-mingw32/bin/
```

### 4\. Cross-Compile with Go

```bash
# Environment Configuration
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1

# Vulkan paths for MinGW
export CGO_CFLAGS="-I/usr/x86_64-w64-mingw32/include"
export CGO_LDFLAGS="-L/usr/x86_64-w64-mingw32/lib -lvulkan-1"

# Build
cd src
go build -ldflags="-s -w" -tags="editor" -o ../kaiju-editor.exe ./
```

### 5\. CI/CD Integration

```yaml
- name: Install MinGW and Cross-Compile Tools
  run: |
    sudo apt-get update
    sudo apt-get install -y mingw-w64 g++-mingw-w64-x86-64

- name: Cache Vulkan SDK for MinGW
  uses: actions/cache@v4
  with:
    path: /usr/x86_64-w64-mingw32
    key: mingw-vulkan-sdk-1.3.290

- name: Install Vulkan Headers and Libraries
  run: |
    # Clone headers
    git clone --depth 1 --branch sdk-1.3.290 \
      https://github.com/KhronosGroup/Vulkan-Headers.git
    cd Vulkan-Headers
    mkdir build && cd build
    cmake -DCMAKE_INSTALL_PREFIX=/usr/x86_64-w64-mingw32 ..
    sudo make install

    # Download Windows Vulkan SDK
    wget https://sdk.lunarg.com/sdk/download/1.3.290.0/windows/VulkanSDK-1.3.290.0-Installer.exe
    7z x VulkanSDK-1.3.290.0-Installer.exe
    sudo cp Lib/vulkan-1.lib /usr/x86_64-w64-mingw32/lib/

- name: Cross-Compile for Windows
  working-directory: ./src
  env:
    CC: x86_64-w64-mingw32-gcc
    CXX: x86_64-w64-mingw32-g++
    GOOS: windows
    GOARCH: amd64
    CGO_ENABLED: 1
    CGO_CFLAGS: -I/usr/x86_64-w64-mingw32/include
    CGO_LDFLAGS: -L/usr/x86_64-w64-mingw32/lib -lvulkan-1
  run: |
    go build -ldflags="-s -w" -tags="editor" -o ../kaiju-editor.exe ./
```

-----

## ⚠️ Potential Challenges

### 1\. Windows System Dependencies

Kaiju uses several libraries that might behave differently:

  * **Vulkan**: Should work with Windows headers/libs
  * **Audio (WASAPI)**: May require Windows SDK headers
  * **Window management**: May require Win32 headers

### 2\. Missing Libraries

MinGW might require additional libraries:

```bash
sudo apt-get install -y \
  mingw-w64-tools \
  wine64  # To test Windows binaries
```

### 3\. Generating `.lib` from `.dll`

If `vulkan-1.lib` is not available:

```bash
# Extract symbols from vulkan-1.dll
dlltool -d vulkan-1.def -l libvulkan-1.a vulkan-1.dll
```

-----

## 🧪 Validation Tests

### Local Test with Wine

```bash
# Install Wine for testing
sudo apt-get install wine64

# Test the Windows binary
wine64 kaiju-editor.exe --version
```

### Test on Real Windows

  * Use a Windows runner just to test the compiled binary
  * Or download and test manually

-----

## 📚 Resources

### Documentation

  * [Vulkan-Headers GitHub](https://github.com/KhronosGroup/Vulkan-Headers)
  * [MSYS2 MinGW Packages](https://packages.msys2.org/packages/mingw-w64-x86_64-vulkan-loader)
  * [LunarG Vulkan SDK](https://vulkan.lunarg.com/)
  * [MinGW-w64 Documentation](http://mingw-w64.org/)

### Articles

  * [Stack Overflow: MinGW with Vulkan](https://stackoverflow.com/questions/35529246/how-do-i-use-vulkan-with-mingw-r-x86-64-32-error)
  * [Conan Cross-Compilation Guide](https://docs.conan.io/2/examples/cross_build/linux_to_windows_mingw.html)
  * [GitHub: MINGW-packages Vulkan](https://github.com/msys2/MINGW-packages/tree/master/mingw-w64-vulkan-headers)

-----

## 📝 Current Status

**Status**: 🚧 **TODO - Not implemented**

**Creation Date**: 2025-12-07

**Priority**: Medium (optimization, not critical)

**Assigned to**: To be determined

## Notes

  * This approach would reduce build time by \~30-40%
  * Would reduce CI costs (Windows minutes = 2x Linux)
  * Requires in-depth validation on real Windows
  * Can be implemented after the current CI stabilization

Would you like me to elaborate on any specific step or provide more detail on the potential challenges?