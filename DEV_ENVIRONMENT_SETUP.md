# Development Environment Setup

## Prerequisites

To develop TablePro, you'll need the following tools installed on your system:

### System Requirements
- C++20 compliant compiler (GCC 10+, Clang 12+, MSVC 2019+)
- CMake 3.24 or later
- Git
- A supported operating system (Linux, macOS, Windows)

### Required Tools
- **Build System**: CMake 3.24+
- **Package Manager**: vcpkg (for dependency management)
- **Compiler**: GCC, Clang, or MSVC depending on platform
- **IDE Support**: Qt Creator, Visual Studio, or any C++ IDE

## Setup Instructions

### 1. Clone the Repository
```bash
git clone https://github.com/your-username/tablepro.git
cd tablepro
```

### 2. Install Dependencies with vcpkg

#### Install vcpkg
```bash
git clone https://github.com/Microsoft/vcpkg.git
cd vcpkg
./bootstrap-vcpkg.sh  # On Linux/macOS
# .\bootstrap-vcpkg.bat # On Windows
./vcpkg integrate install
```

#### Install Required Libraries

*Note: TablePro uses vcpkg in manifest mode (`vcpkg.json`). Dependencies will be automatically installed when generating the CMake project in the next step.*

### 3. Build the Project

#### Configure the Project
```bash
cmake -B build -DCMAKE_TOOLCHAIN_FILE=/path/to/vcpkg/scripts/buildsystems/vcpkg.cmake -DCMAKE_BUILD_TYPE=Debug
```

#### Compile
```bash
cmake --build build --parallel
```

### 4. Running the Application
```bash
./build/bin/TablePro
```

## Development Guidelines

### Code Style
- Use clang-format with the provided .clang-format file
- Follow Qt naming conventions
- Use RAII for resource management
- Prefer smart pointers over raw pointers

### Git Workflow
- Create feature branches for new features
- Keep commits atomic and well-described
- Run pre-commit hooks before pushing
- Follow conventional commit messages

## Troubleshooting

### Common Issues
- **Missing Qt dependencies**: Ensure vcpkg is properly configured and Qt dependencies are installed
- **Linking errors**: Make sure all dependencies are found during CMake configuration
- **Platform-specific issues**: Check platform-specific vcpkg triplet configuration

### Debugging Tips
- Use `cmake --build build --verbose` for detailed build output
- Check CMake configuration messages for missing dependencies
- Ensure vcpkg is integrated into your development environment