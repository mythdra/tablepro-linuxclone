# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TablePro is a **C++20 + Qt 6.6 LTS** cross-platform database client targeting macOS, Windows, and Linux as native binaries. The project is currently in the planning/specification phase with comprehensive specs in `/specs/` and implementation phases in `/plans/phases/`.

**Stack:**
- Language: C++20
- GUI Framework: Qt 6.6 LTS (Qt Widgets, not QML)
- Build System: CMake 3.24+ + vcpkg
- Testing: Qt Test + Catch2

## Build Commands

```bash
# Prerequisites (macOS)
brew install qt@6 vcpkg cmake ninja
export VCPKG_ROOT=/path/to/vcpkg

# Configure (Debug)
cmake -B build \
  -DCMAKE_TOOLCHAIN_FILE=$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake \
  -DCMAKE_BUILD_TYPE=Debug \
  -GNinja

# Build
cmake --build build -j$(nproc)

# Run tests
ctest --test-dir build --output-on-failure

# Run specific test
ctest -R DriverTest --verbose

# Format code
clang-format -i src/**/*.cpp src/**/*.hpp
```

## Architecture

```
src/
├── core/           # Business logic (no Qt GUI dependencies)
│   ├── DatabaseDriver.hpp    # Abstract driver interface
│   ├── PostgresDriver.cpp    # libpq implementation
│   ├── MysqlDriver.cpp       # libmysql implementation
│   ├── ConnectionManager.cpp
│   ├── QueryExecutor.cpp
│   ├── ChangeTracker.cpp
│   └── SqlGenerator.cpp
├── ui/             # Qt UI components
│   ├── MainWindow.cpp
│   ├── ConnectionDialog.cpp
│   ├── DataGrid/   # QTableView + custom model
│   └── Editor/     # QScintilla SQL editor
├── services/       # Application services
│   ├── ExportService.cpp
│   ├── ImportService.cpp
│   └── HistoryService.cpp
└── main.cpp
```

**Communication Pattern:** UI → Core via direct method calls; Core → UI via Qt signals/slots.

## C++ Coding Conventions

### Naming
- **Classes**: PascalCase (`DatabaseConnection`, `QueryResult`)
- **Methods**: camelCase (`executeQuery`, `connectToDatabase`)
- **Member variables**: `m_` prefix (`m_connection`, `m_queryCache`)
- **Constants**: `kPascalCase` (`kDefaultTimeout`)
- **Enums**: PascalCase with `k` prefix (`ConnectionState::kConnected`)

### Memory Management
- **Qt objects**: Parent-child ownership (parent deletes children)
- **Non-Qt objects**: `std::unique_ptr` by default, `std::shared_ptr` for shared ownership
- **Raw pointers**: Non-owning references only

### Key Patterns
```cpp
// RAII always - no manual cleanup
std::unique_ptr<DatabaseDriver> m_driver;

// Qt parent-child for UI
auto* button = new QPushButton(this);  // this takes ownership

// Never store passwords in structs - use QKeychain
// Use signals/slots for async communication
// Q_OBJECT macro required for classes with signals/slots
```

## Pre-commit Checklist
- [ ] `clang-format` applied
- [ ] `cmake --build` succeeds with no warnings (`-Wall -Wextra -Wpedantic`)
- [ ] `ctest` passes
- [ ] Qt slots/signals properly connected
- [ ] No raw `new` without ownership (use smart pointers or Qt parent)

## Key Documentation
- **Specifications**: `/specs/` - detailed feature and technical specs
- **Implementation Phases**: `/plans/phases/` - 18 phases with 400+ tasks
- **Reference**: `/plans/reference/` - architecture, conventions, dependencies