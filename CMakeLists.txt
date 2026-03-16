cmake_minimum_required(VERSION 3.24)

project(TablePro VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Set output directories
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)

# Compiler-specific options
if(CMAKE_CXX_COMPILER_ID STREQUAL "GNU" OR CMAKE_CXX_COMPILER_ID MATCHES "Clang")
    add_compile_options(-Wall -Wextra -Wpedantic)
elseif(MSVC)
    add_compile_options(/W4)
endif()

# Find Qt6
find_package(Qt6 REQUIRED COMPONENTS Core Widgets Sql PrintSupport Concurrent Test Network)

# Enable AUTOMOC, AUTORCC, AUTOUIC
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

# Resources
set(TABLEPRO_RESOURCES
    ${CMAKE_CURRENT_SOURCE_DIR}/resources/resources.qrc
)

# Include directories
set(TABLEPRO_INCLUDE_DIRS
    ${CMAKE_CURRENT_SOURCE_DIR}/src
    ${CMAKE_CURRENT_SOURCE_DIR}/include
)

# Add subdirectories
add_subdirectory(src/core)
add_subdirectory(src/ui)
add_subdirectory(src/services)
add_subdirectory(src/drivers)

# Main application source
set(TABLEPRO_MAIN_SOURCES
    src/main.cpp
    ${TABLEPRO_RESOURCES}
)

# Add executable
add_executable(${PROJECT_NAME} ${TABLEPRO_MAIN_SOURCES})

# Link Qt libraries
target_link_libraries(${PROJECT_NAME}
    Qt6::Core
    Qt6::Widgets
    Qt6::Sql
    tablepro_core
    tablepro_ui
    tablepro_services
    tablepro_drivers
)

# Include directories
target_include_directories(${PROJECT_NAME} PRIVATE ${TABLEPRO_INCLUDE_DIRS})

# Set application properties
set_target_properties(${PROJECT_NAME} PROPERTIES
    WIN32_EXECUTABLE TRUE
    MACOSX_BUNDLE TRUE
)

# Enable Qt Automoc, Autorcc, and Uic
qt6_standard_project_setup()

# Enable testing
enable_testing()

# Add test subdirectory
add_subdirectory(tests)