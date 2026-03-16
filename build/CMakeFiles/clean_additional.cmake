# Additional clean files
cmake_minimum_required(VERSION 3.16)

if("${CONFIG}" STREQUAL "" OR "${CONFIG}" STREQUAL "Debug")
  file(REMOVE_RECURSE
  "CMakeFiles/TablePro_autogen.dir/AutogenUsed.txt"
  "CMakeFiles/TablePro_autogen.dir/ParseCache.txt"
  "TablePro_autogen"
  "src/core/CMakeFiles/tablepro_core_autogen.dir/AutogenUsed.txt"
  "src/core/CMakeFiles/tablepro_core_autogen.dir/ParseCache.txt"
  "src/core/tablepro_core_autogen"
  "src/drivers/CMakeFiles/tablepro_drivers_autogen.dir/AutogenUsed.txt"
  "src/drivers/CMakeFiles/tablepro_drivers_autogen.dir/ParseCache.txt"
  "src/drivers/tablepro_drivers_autogen"
  "src/services/CMakeFiles/tablepro_services_autogen.dir/AutogenUsed.txt"
  "src/services/CMakeFiles/tablepro_services_autogen.dir/ParseCache.txt"
  "src/services/tablepro_services_autogen"
  "src/ui/CMakeFiles/tablepro_ui_autogen.dir/AutogenUsed.txt"
  "src/ui/CMakeFiles/tablepro_ui_autogen.dir/ParseCache.txt"
  "src/ui/tablepro_ui_autogen"
  "tests/integration/CMakeFiles/test_postgres_integration_autogen.dir/AutogenUsed.txt"
  "tests/integration/CMakeFiles/test_postgres_integration_autogen.dir/ParseCache.txt"
  "tests/integration/test_postgres_integration_autogen"
  "tests/unit/CMakeFiles/test_connection_manager_autogen.dir/AutogenUsed.txt"
  "tests/unit/CMakeFiles/test_connection_manager_autogen.dir/ParseCache.txt"
  "tests/unit/CMakeFiles/test_database_driver_autogen.dir/AutogenUsed.txt"
  "tests/unit/CMakeFiles/test_database_driver_autogen.dir/ParseCache.txt"
  "tests/unit/CMakeFiles/test_postgres_driver_autogen.dir/AutogenUsed.txt"
  "tests/unit/CMakeFiles/test_postgres_driver_autogen.dir/ParseCache.txt"
  "tests/unit/test_connection_manager_autogen"
  "tests/unit/test_database_driver_autogen"
  "tests/unit/test_postgres_driver_autogen"
  )
endif()
