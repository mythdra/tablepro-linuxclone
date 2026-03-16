# CMakeLists.txt for src/services
set(TABLEPRO_SERVICES_SOURCES
    services.cpp
)

add_library(tablepro_services STATIC ${TABLEPRO_SERVICES_SOURCES})
target_include_directories(tablepro_services PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/../../include/services)