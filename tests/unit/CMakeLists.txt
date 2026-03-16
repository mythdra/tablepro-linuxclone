# CMakeLists.txt for unit tests

# Test for DatabaseDriver and related components
set(TEST_DATABASE_DRIVER_SOURCES
    test_database_driver.cpp
)

add_executable(test_database_driver ${TEST_DATABASE_DRIVER_SOURCES})
target_link_libraries(test_database_driver Qt6::Core Qt6::Test tablepro_core)
target_include_directories(test_database_driver PRIVATE ${CMAKE_SOURCE_DIR}/include)

add_test(NAME test_database_driver COMMAND test_database_driver)

# Test for ConnectionManager
set(TEST_CONNECTION_MANAGER_SOURCES
    test_connection_manager.cpp
)

add_executable(test_connection_manager ${TEST_CONNECTION_MANAGER_SOURCES})
target_link_libraries(test_connection_manager Qt6::Core Qt6::Test tablepro_core)
target_include_directories(test_connection_manager PRIVATE ${CMAKE_SOURCE_DIR}/include)

add_test(NAME test_connection_manager COMMAND test_connection_manager)

# Test for PostgresDriver
set(TEST_POSTGRES_DRIVER_SOURCES
    test_postgres_driver.cpp
)

add_executable(test_postgres_driver ${TEST_POSTGRES_DRIVER_SOURCES})
target_link_libraries(test_postgres_driver Qt6::Core Qt6::Test tablepro_core tablepro_drivers)
target_include_directories(test_postgres_driver PRIVATE ${CMAKE_SOURCE_DIR}/include)

add_test(NAME test_postgres_driver COMMAND test_postgres_driver)