# CMakeLists.txt for src/core
set(TABLEPRO_CORE_SOURCES
    database_driver.cpp
    database_exceptions.cpp
    connection_manager.cpp
    query_executor.cpp
    change_tracker.cpp
    sql_generator.cpp
)

set(TABLEPRO_CORE_HEADERS
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/core/database_types.h
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/core/database_driver.h
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/core/database_exceptions.h
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/core/connection_config.h
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/core/connection_manager.h
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/core/query_result.h
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/core/query_executor.h
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/core/change_tracker.h
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/core/sql_generator.h
)

add_library(tablepro_core STATIC ${TABLEPRO_CORE_SOURCES} ${TABLEPRO_CORE_HEADERS})
target_include_directories(tablepro_core PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/../../include/core)
target_link_libraries(tablepro_core Qt6::Core Qt6::Sql)