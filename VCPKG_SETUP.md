# VCPKG Setup Process

## Prerequisites

Before building TablePro, you'll need to install vcpkg and set up the required dependencies.

## Installing vcpkg

1. Clone the vcpkg repository:
   ```bash
   git clone https://github.com/Microsoft/vcpkg.git
   cd vcpkg
   ```

2. Bootstrap vcpkg:
   ```bash
   ./bootstrap-vcpkg.sh  # On Linux/macOS
   # or
   .\bootstrap-vcpkg.bat # On Windows
   ```

3. Integrate vcpkg with your system:
   ```bash
   ./vcpkg integrate install
   ```

## Installing Dependencies

TablePro uses the following dependencies managed by vcpkg. Install them using:

```bash
./vcpkg install qtbase qttools qtscintilla libpq openssl zlib libmysql sqlite3
```

## Building with vcpkg

Once dependencies are installed, configure your project to use vcpkg:

```bash
cmake -B build -DCMAKE_TOOLCHAIN_FILE=[vcpkg root]/scripts/buildsystems/vcpkg.cmake
cmake --build build
```

Replace `[vcpkg root]` with the path to your vcpkg installation.