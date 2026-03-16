# CMakeLists.txt for src/drivers
set(CMAKE_AUTOMOC ON)

set(TABLEPRO_DRIVERS_SOURCES
    drivers.cpp
    postgres_driver.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/drivers/postgres_driver.h
)

add_library(tablepro_drivers STATIC ${TABLEPRO_DRIVERS_SOURCES})
target_include_directories(tablepro_drivers PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/../../include/drivers)
target_include_directories(tablepro_drivers PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/../../include)

# Find PostgreSQL - try vcpkg first, then Homebrew
find_package(PostgreSQL QUIET)

if(NOT PostgreSQL_FOUND)
    # Try Homebrew path on macOS
    if(APPLE)
        execute_process(
            COMMAND brew --prefix libpq
            OUTPUT_VARIABLE HOMEBREW_LIBPQ_PATH
            OUTPUT_STRIP_TRAILING_WHITESPACE
            RESULT_VARIABLE BREW_RESULT
        )
        if(BREW_RESULT EQUAL 0)
            set(PostgreSQL_INCLUDE_DIRS "${HOMEBREW_LIBPQ_PATH}/include")
            set(PostgreSQL_LIBRARIES "${HOMEBREW_LIBPQ_PATH}/lib/libpq.dylib")
            message(STATUS "Found PostgreSQL via Homebrew: ${HOMEBREW_LIBPQ_PATH}")
        endif()
    endif()
endif()

target_include_directories(tablepro_drivers PUBLIC ${PostgreSQL_INCLUDE_DIRS})
target_link_libraries(tablepro_drivers PUBLIC Qt6::Core Qt6::Sql tablepro_core ${PostgreSQL_LIBRARIES})