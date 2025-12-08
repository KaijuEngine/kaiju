# Cross-Compilation Windows depuis Linux (TODO)

## Objectif
Éliminer le runner Windows du CI en faisant la cross-compilation pour Windows depuis Linux avec MinGW-w64.

## Avantages
- ✅ Tout builder sur le même runner Linux (plus rapide)
- ✅ Moins de complexité (pas besoin de gérer 2 OS différents)
- ✅ Moins cher (les runners Windows coûtent 2x plus en minutes CI)
- ✅ Plus fiable (Linux est plus stable pour CI)

## Approche Recommandée

### 1. Installer MinGW sur Ubuntu
```bash
sudo apt-get update
sudo apt-get install -y \
  mingw-w64 \
  g++-mingw-w64-x86-64 \
  gcc-mingw-w64-x86-64
```

### 2. Obtenir les Headers Vulkan

**Option A - Cloner depuis GitHub (recommandé)**:
```bash
# Les headers Vulkan sont platform-agnostic
git clone --depth 1 --branch sdk-1.3.290 \
  https://github.com/KhronosGroup/Vulkan-Headers.git
cd Vulkan-Headers
mkdir build && cd build

# Installer dans le préfixe MinGW
cmake -DCMAKE_INSTALL_PREFIX=/usr/x86_64-w64-mingw32 ..
sudo make install
```

**Option B - Copier headers système**:
```bash
# Les headers Vulkan Linux fonctionnent aussi pour Windows
sudo apt-get install libvulkan-dev
sudo cp -r /usr/include/vulkan /usr/x86_64-w64-mingw32/include/
sudo cp -r /usr/include/vk_video /usr/x86_64-w64-mingw32/include/
```

### 3. Obtenir la Bibliothèque Vulkan Windows

**Option A - Télécharger SDK Windows**:
```bash
wget https://sdk.lunarg.com/sdk/download/1.3.290.0/windows/VulkanSDK-1.3.290.0-Installer.exe
7z x VulkanSDK-1.3.290.0-Installer.exe
sudo cp Lib/vulkan-1.lib /usr/x86_64-w64-mingw32/lib/
```

**Option B - Extraire depuis MSYS2** (plus complexe):
```bash
# Télécharger le paquet MSYS2
wget https://repo.msys2.org/mingw/mingw64/mingw-w64-x86_64-vulkan-loader-*.pkg.tar.zst
tar -I zstd -xf mingw-w64-x86_64-vulkan-loader-*.pkg.tar.zst
sudo cp mingw64/lib/libvulkan.a /usr/x86_64-w64-mingw32/lib/
sudo cp mingw64/bin/vulkan-1.dll /usr/x86_64-w64-mingw32/bin/
```

### 4. Cross-Compiler avec Go

```bash
# Configuration de l'environnement
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1

# Chemins Vulkan pour MinGW
export CGO_CFLAGS="-I/usr/x86_64-w64-mingw32/include"
export CGO_LDFLAGS="-L/usr/x86_64-w64-mingw32/lib -lvulkan-1"

# Build
cd src
go build -ldflags="-s -w" -tags="editor" -o ../kaiju-editor.exe ./
```

### 5. Intégration CI/CD

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

## Défis Potentiels

### 1. Dépendances Système Windows
Kaiju utilise plusieurs bibliothèques qui peuvent avoir des comportements différents:
- **Vulkan**: Devrait fonctionner avec les headers/libs Windows
- **Audio (WASAPI)**: Peut nécessiter des headers Windows SDK
- **Window management**: Peut nécessiter des headers Win32

### 2. Bibliothèques Manquantes
MinGW peut nécessiter des bibliothèques supplémentaires:
```bash
sudo apt-get install -y \
  mingw-w64-tools \
  wine64  # Pour tester les binaires Windows
```

### 3. Génération de `.lib` depuis `.dll`
Si vulkan-1.lib n'est pas disponible:
```bash
# Extraire les symboles de vulkan-1.dll
dlltool -d vulkan-1.def -l libvulkan-1.a vulkan-1.dll
```

## Tests de Validation

### Test Local avec Wine
```bash
# Installer Wine pour tester
sudo apt-get install wine64

# Tester le binaire Windows
wine64 kaiju-editor.exe --version
```

### Test sur Windows Réel
- Utiliser un runner Windows juste pour tester le binaire compilé
- Ou télécharger et tester manuellement

## Ressources

### Documentation
- [Vulkan-Headers GitHub](https://github.com/KhronosGroup/Vulkan-Headers)
- [MSYS2 MinGW Packages](https://packages.msys2.org/packages/mingw-w64-x86_64-vulkan-headers)
- [LunarG Vulkan SDK](https://vulkan.lunarg.com/)
- [MinGW-w64 Documentation](http://mingw-w64.org/)

### Articles
- [Stack Overflow: MinGW with Vulkan](https://stackoverflow.com/questions/35529246/how-do-i-use-vulkan-with-mingw-r-x86-64-32-error)
- [Conan Cross-Compilation Guide](https://docs.conan.io/2/examples/cross_build/linux_to_windows_mingw.html)
- [GitHub: MINGW-packages Vulkan](https://github.com/msys2/MINGW-packages/tree/master/mingw-w64-vulkan-headers)

## État Actuel

**Status**: 🚧 TODO - Non implémenté

**Date de création**: 2025-12-07

**Priorité**: Moyenne (optimisation, pas critique)

**Assigné à**: À déterminer

## Notes
- Cette approche réduirait le temps de build de ~30-40%
- Réduirait les coûts CI (minutes Windows = 2x Linux)
- Nécessite validation approfondie sur Windows réel
- Peut être implémenté après stabilisation du CI actuel
