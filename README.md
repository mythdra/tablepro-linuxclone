# TablePro - Cross-Platform Database Client

TablePro is a modern, cross-platform database client supporting PostgreSQL, MySQL, SQLite, and other database systems. Built with C++20 and Qt 6.6 LTS for maximum performance and native look-and-feel across platforms.

## Features

- Support for multiple database systems (PostgreSQL, MySQL, SQLite, and more)
- Intuitive UI with schema browser and data grid
- Powerful SQL editor with syntax highlighting
- Data import/export capabilities
- SSH tunneling and SSL support
- Cross-platform (Windows, macOS, Linux)

## Requirements

- C++20 compiler
- CMake 3.24+
- Qt 6.6 LTS
- vcpkg (for dependency management)

## Building from Source

### Prerequisites

1. Install CMake 3.24 or later
2. Install a C++20 compliant compiler
3. Install Git
4. Install vcpkg for dependency management

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/your-repo/tablepro.git
   cd tablepro
   ```

2. Install dependencies with vcpkg:
   ```bash
   # Install vcpkg if you haven't already
   git clone https://github.com/Microsoft/vcpkg.git
   cd vcpkg
   ./bootstrap-vcpkg.sh  # On Linux/macOS
   # .\bootstrap-vcpkg.bat # On Windows
   ./vcpkg integrate install

   # Install required libraries
   ./vcpkg install qtbase qttools qtscintilla libpq openssl zlib libmysql sqlite3
   ```

3. Build the project:
   ```bash
   cmake -B build -DCMAKE_TOOLCHAIN_FILE=/path/to/vcpkg/scripts/buildsystems/vcpkg.cmake -DCMAKE_BUILD_TYPE=Debug
   cmake --build build --parallel
   ```

4. Run the application:
   ```bash
   ./build/bin/TablePro
   ```

## Development

See [DEV_ENVIRONMENT_SETUP.md](DEV_ENVIRONMENT_SETUP.md) for detailed development environment setup instructions.

## Contributing

We welcome contributions! Please see our contributing guidelines for more information.

## License

MIT License. See LICENSE file for details.