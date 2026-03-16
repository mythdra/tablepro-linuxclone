# CMakeLists.txt for src/ui
set(CMAKE_AUTOMOC ON)

set(TABLEPRO_UI_SOURCES
    mainwindow.cpp
    schema_tree_view.cpp
    connection_dialog.cpp
    theme_manager.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/ui/mainwindow.h
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/ui/schema_tree_view.h
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/ui/connection_dialog.h
    ${CMAKE_CURRENT_SOURCE_DIR}/../../include/ui/theme_manager.h
)

add_library(tablepro_ui STATIC ${TABLEPRO_UI_SOURCES})
target_include_directories(tablepro_ui PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/../../include/ui)
target_include_directories(tablepro_ui PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/../../include)

# Link Qt libraries to UI library
target_link_libraries(tablepro_ui Qt6::Core Qt6::Widgets tablepro_core tablepro_drivers)