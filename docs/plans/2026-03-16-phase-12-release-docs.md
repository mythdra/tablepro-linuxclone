# Phase 12: Release & Docs Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task with proper verification at each step.

**Goal:** Configure cross-platform builds, code signing, installers, and create documentation for v1.0.0 release.

**Architecture:** CMake presets for each platform. Native installers (DMG, NSIS, AppImage). Doxygen for API docs. User guide in Markdown.

**Tech Stack:** CMake, codesign (macOS), signtool (Windows), linuxdeploy, Doxygen

---

## Task 1: macOS Build Configuration

**Files:**
- Create: `build/darwin/Info.plist`
- Create: `build/darwin/entitlements.xml`

**Step 1: Create Info.plist**

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>TablePro</string>
    <key>CFBundleIdentifier</key>
    <string>app.tablepro.TablePro</string>
    <key>CFBundleVersion</key>
    <string>1.0.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundleExecutable</key>
    <string>tablepro</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon.icns</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>LSMinimumSystemVersion</key>
    <string>12.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2024 TablePro. All rights reserved.</string>
</dict>
</plist>
```

**Step 2: Create entitlements.xml**

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <key>com.apple.security.cs.disable-library-validation</key>
    <true/>
    <key>com.apple.security.files.user-selected.read-write</key>
    <true/>
    <key>com.apple.security.network.client</key>
    <true/>
</dict>
</plist>
```

**Step 3: Update CMakeLists.txt for macOS bundle**

```cmake
if(APPLE)
    set_target_properties(tablepro PROPERTIES
        MACOSX_BUNDLE TRUE
        MACOSX_BUNDLE_INFO_PLIST "${CMAKE_SOURCE_DIR}/build/darwin/Info.plist"
    )
endif()
```

**Commit:**

```bash
git add build/darwin/
git commit -m "build: Add macOS bundle configuration"
```

---

## Task 2: Windows Build Configuration

**Files:**
- Create: `build/windows/tablepro.rc`
- Create: `build/windows/installer.nsi`

**Step 1: Create Windows resource file**

```rc
#include <windows.h>

VS_VERSION_INFO VERSIONINFO
FILEVERSION 1,0,0,0
PRODUCTVERSION 1,0,0,0
FILEFLAGSMASK 0x3fL
FILEFLAGS 0x0L
FILEOS 0x40004L
FILETYPE 0x1L
FILESUBTYPE 0x0L
BEGIN
    BLOCK "StringFileInfo"
    BEGIN
        BLOCK "040904E4"
        BEGIN
            VALUE "CompanyName", "TablePro"
            VALUE "FileDescription", "TablePro Database Client"
            VALUE "FileVersion", "1.0.0.0"
            VALUE "ProductName", "TablePro"
            VALUE "ProductVersion", "1.0.0.0"
        END
    END
END

1 ICON "tablepro.ico"
```

**Step 2: Create NSIS installer script**

```nsis
!include "MUI2.nsh"

Name "TablePro"
OutFile "TablePro-1.0.0-Setup.exe"
InstallDir "$PROGRAMFILES64\TablePro"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_LANGUAGE "English"

Section "Install"
    SetOutPath "$INSTDIR"
    File /r "build\windows-release\*.*"

    CreateDirectory "$SMPROGRAMS\TablePro"
    CreateShortCut "$SMPROGRAMS\TablePro\TablePro.lnk" "$INSTDIR\tablepro.exe"

    WriteUninstaller "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Uninstall"
    Delete "$INSTDIR\*.*"
    RMDir "$INSTDIR"
    Delete "$SMPROGRAMS\TablePro\TablePro.lnk"
    RMDir "$SMPROGRAMS\TablePro"
SectionEnd
```

**Commit:**

```bash
git add build/windows/
git commit -m "build: Add Windows installer configuration"
```

---

## Task 3: Linux Build Configuration

**Files:**
- Create: `build/linux/tablepro.desktop`
- Create: `build/linux/AppImageBuilder.yml`

**Step 1: Create .desktop file**

```desktop
[Desktop Entry]
Name=TablePro
Comment=Modern Database Client
Exec=tablepro %f
Icon=tablepro
Terminal=false
Type=Application
Categories=Development;Database;
MimeType=text/x-sql;
Keywords=database;sql;query;
```

**Step 2: Create AppImage builder config**

```yaml
version: 1
script:
  - cmake -B build -DCMAKE_BUILD_TYPE=Release
  - cmake --build build
AppDir:
  path: ./AppDir
  app_info:
    id: app.tablepro.TablePro
    name: TablePro
    icon: tablepro
    version: 1.0.0
    exec: usr/bin/tablepro
  files:
    include:
      - usr/lib/**/*.so*
    exclude:
      - usr/share/man
      - usr/share/doc
AppImage:
  arch: x86_64
```

**Commit:**

```bash
git add build/linux/
git commit -m "build: Add Linux AppImage configuration"
```

---

## Task 4: GitHub Actions Release Workflow

**Files:**
- Create: `.github/workflows/release.yml`

**Step 1: Create release workflow**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build-macos:
    runs-on: macos-14
    steps:
      - uses: actions/checkout@v4

      - name: Build
        run: |
          cmake -B build -DCMAKE_BUILD_TYPE=Release
          cmake --build build -j$(sysctl -n hw.ncpu)

      - name: Create DMG
        run: |
          brew install create-dmg
          create-dmg TablePro-${{ github.ref_name }}.dmg \
            --volname "TablePro" \
            --volicon "resources/icons/AppIcon.icns" \
            build/tablepro.app

      - uses: actions/upload-artifact@v4
        with:
          name: macos-dmg
          path: "*.dmg"

  build-windows:
    runs-on: windows-2022
    steps:
      - uses: actions/checkout@v4

      - name: Build
        run: |
          cmake -B build -G "Visual Studio 17 2022" -A x64
          cmake --build build --config Release

      - name: Create Installer
        run: |
          makensis build/windows/installer.nsi

      - uses: actions/upload-artifact@v4
        with:
          name: windows-installer
          path: "*.exe"

  build-linux:
    runs-on: ubuntu-22.04
    steps:
      - uses: actions/checkout@v4

      - name: Build
        run: |
          cmake -B build -DCMAKE_BUILD_TYPE=Release
          cmake --build build -j$(nproc)

      - name: Create AppImage
        run: |
          wget https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
          chmod +x linuxdeploy-x86_64.AppImage
          ./linuxdeploy-x86_64.AppImage --appdir AppDir -e build/tablepro -d build/linux/tablepro.desktop -i resources/icons/tablepro.png --output appimage

      - uses: actions/upload-artifact@v4
        with:
          name: linux-appimage
          path: "*.AppImage"

  release:
    needs: [build-macos, build-windows, build-linux]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            macos-dmg/*.dmg
            windows-installer/*.exe
            linux-appimage/*.AppImage
          generate_release_notes: true
```

**Commit:**

```bash
git add .github/workflows/release.yml
git commit -m "ci: Add release workflow for all platforms"
```

---

## Task 5: User Documentation

**Files:**
- Create: `docs/user-guide/getting-started.md`
- Create: `docs/user-guide/connections.md`
- Create: `docs/user-guide/query-editor.md`

**Step 1: Create getting started guide**

```markdown
# Getting Started with TablePro

## Installation

### macOS
1. Download the DMG file
2. Open and drag TablePro to Applications
3. Launch from Applications folder

### Windows
1. Download the installer
2. Run the installer
3. Launch from Start Menu

### Linux
1. Download the AppImage
2. Make executable: `chmod +x TablePro-*.AppImage`
3. Run: `./TablePro-*.AppImage`

## Quick Start

1. Click "New Connection" in the toolbar
2. Enter your database credentials
3. Click "Connect"
4. Browse tables in the sidebar
5. Double-click a table to view data
6. Open a Query tab to write SQL

## Next Steps

- [Managing Connections](connections.md)
- [Using the Query Editor](query-editor.md)
```

**Step 2: Create connections guide**

```markdown
# Managing Connections

## Creating a Connection

1. Click "New Connection" or press Ctrl+N
2. Fill in the connection details:
   - Name: Display name for the connection
   - Type: Database type (PostgreSQL, MySQL, etc.)
   - Host: Database server hostname
   - Port: Database port
   - Database: Database name
   - Username: Database username
3. Click "Test Connection" to verify
4. Click "Save" to store the connection

## Connection Options

### SSH Tunnel
Enable SSH tunnel for secure connections:
- SSH Host: SSH server hostname
- SSH Port: SSH port (default 22)
- SSH Username: SSH username
- SSH Key: Path to private key file

### SSL/TLS
Enable SSL for encrypted connections:
- CA Certificate: Path to CA cert
- Client Certificate: Path to client cert
- Client Key: Path to client key
```

**Commit:**

```bash
git add docs/user-guide/
git commit -m "docs: Add user guide documentation"
```

---

## Task 6: API Documentation

**Files:**
- Create: `Doxyfile`

**Step 1: Create Doxyfile**

```
PROJECT_NAME = "TablePro"
PROJECT_NUMBER = 1.0.0
PROJECT_BRIEF = "Cross-platform database client"

INPUT = src/
RECURSIVE = YES
FILE_PATTERNS = *.hpp *.cpp

EXTRACT_ALL = YES
EXTRACT_PRIVATE = YES
EXTRACT_STATIC = YES

GENERATE_HTML = YES
HTML_OUTPUT = docs/api

GENERATE_LATEX = NO

HAVE_DOT = YES
UML_LOOK = YES
```

**Commit:**

```bash
git add Doxyfile
git commit -m "docs: Add Doxygen configuration"
```

---

## Task 7: Final Verification

**Step 1: Build all platforms**

```bash
# macOS
cmake -B build/macos -DCMAKE_BUILD_TYPE=Release
cmake --build build/macos -j$(sysctl -n hw.ncpu)

# Run app
./build/macos/tablepro.app/Contents/MacOS/tablepro
```

**Step 2: Verify all features**

- [ ] Application launches
- [ ] Can create connections
- [ ] Can connect to PostgreSQL
- [ ] Schema tree shows tables
- [ ] Data grid displays results
- [ ] SQL editor works
- [ ] Export functions
- [ ] Settings persist

**Step 3: Create release commit**

```bash
git add -A
git commit -m "release: v1.0.0

- Complete TablePro database client
- Support for 8 database types
- Cross-platform builds (macOS, Windows, Linux)
- SSH tunneling and SSL support
- Query history with FTS search
- Export/Import services
- License management

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"

git tag -a v1.0.0 -m "TablePro v1.0.0 Release"
git push origin main --tags
```

---

## Acceptance Criteria

- [ ] macOS DMG builds and runs
- [ ] Windows installer builds and runs
- [ ] Linux AppImage builds and runs
- [ ] GitHub release workflow passes
- [ ] User documentation complete
- [ ] API documentation generates
- [ ] All core features functional
- [ ] v1.0.0 tag created

---

**Phase 12 Complete.**

🎉 **TablePro v1.0.0 Ready for Release!**