/****************************************************************************
** Meta object code from reading C++ file 'change_tracker.h'
**
** Created by: The Qt Meta Object Compiler version 69 (Qt 6.10.2)
**
** WARNING! All changes made in this file will be lost!
*****************************************************************************/

#include "../../../../../include/core/change_tracker.h"
#include <QtCore/qmetatype.h>

#include <QtCore/qtmochelpers.h>

#include <memory>


#include <QtCore/qxptype_traits.h>
#if !defined(Q_MOC_OUTPUT_REVISION)
#error "The header file 'change_tracker.h' doesn't include <QObject>."
#elif Q_MOC_OUTPUT_REVISION != 69
#error "This file was generated using the moc from 6.10.2. It"
#error "cannot be used with the include files from this version of Qt."
#error "(The moc has changed too much.)"
#endif

#ifndef Q_CONSTINIT
#define Q_CONSTINIT
#endif

QT_WARNING_PUSH
QT_WARNING_DISABLE_DEPRECATED
QT_WARNING_DISABLE_GCC("-Wuseless-cast")
namespace {
struct qt_meta_tag_ZN8tablepro13ChangeTrackerE_t {};
} // unnamed namespace

template <> constexpr inline auto tablepro::ChangeTracker::qt_create_metaobjectdata<qt_meta_tag_ZN8tablepro13ChangeTrackerE_t>()
{
    namespace QMC = QtMocConstants;
    QtMocHelpers::StringRefStorage qt_stringData {
        "tablepro::ChangeTracker",
        "changeRecorded",
        "",
        "ChangeRecord",
        "change",
        "changesApplied",
        "count",
        "changesDiscarded",
        "changeUndone",
        "changeRedone",
        "dirtyStateChanged",
        "isDirty"
    };

    QtMocHelpers::UintData qt_methods {
        // Signal 'changeRecorded'
        QtMocHelpers::SignalData<void(const ChangeRecord &)>(1, 2, QMC::AccessPublic, QMetaType::Void, {{
            { 0x80000000 | 3, 4 },
        }}),
        // Signal 'changesApplied'
        QtMocHelpers::SignalData<void(int)>(5, 2, QMC::AccessPublic, QMetaType::Void, {{
            { QMetaType::Int, 6 },
        }}),
        // Signal 'changesDiscarded'
        QtMocHelpers::SignalData<void(int)>(7, 2, QMC::AccessPublic, QMetaType::Void, {{
            { QMetaType::Int, 6 },
        }}),
        // Signal 'changeUndone'
        QtMocHelpers::SignalData<void(const ChangeRecord &)>(8, 2, QMC::AccessPublic, QMetaType::Void, {{
            { 0x80000000 | 3, 4 },
        }}),
        // Signal 'changeRedone'
        QtMocHelpers::SignalData<void(const ChangeRecord &)>(9, 2, QMC::AccessPublic, QMetaType::Void, {{
            { 0x80000000 | 3, 4 },
        }}),
        // Signal 'dirtyStateChanged'
        QtMocHelpers::SignalData<void(bool)>(10, 2, QMC::AccessPublic, QMetaType::Void, {{
            { QMetaType::Bool, 11 },
        }}),
    };
    QtMocHelpers::UintData qt_properties {
    };
    QtMocHelpers::UintData qt_enums {
    };
    return QtMocHelpers::metaObjectData<ChangeTracker, qt_meta_tag_ZN8tablepro13ChangeTrackerE_t>(QMC::MetaObjectFlag{}, qt_stringData,
            qt_methods, qt_properties, qt_enums);
}
Q_CONSTINIT const QMetaObject tablepro::ChangeTracker::staticMetaObject = { {
    QMetaObject::SuperData::link<QObject::staticMetaObject>(),
    qt_staticMetaObjectStaticContent<qt_meta_tag_ZN8tablepro13ChangeTrackerE_t>.stringdata,
    qt_staticMetaObjectStaticContent<qt_meta_tag_ZN8tablepro13ChangeTrackerE_t>.data,
    qt_static_metacall,
    nullptr,
    qt_staticMetaObjectRelocatingContent<qt_meta_tag_ZN8tablepro13ChangeTrackerE_t>.metaTypes,
    nullptr
} };

void tablepro::ChangeTracker::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    auto *_t = static_cast<ChangeTracker *>(_o);
    if (_c == QMetaObject::InvokeMetaMethod) {
        switch (_id) {
        case 0: _t->changeRecorded((*reinterpret_cast<std::add_pointer_t<ChangeRecord>>(_a[1]))); break;
        case 1: _t->changesApplied((*reinterpret_cast<std::add_pointer_t<int>>(_a[1]))); break;
        case 2: _t->changesDiscarded((*reinterpret_cast<std::add_pointer_t<int>>(_a[1]))); break;
        case 3: _t->changeUndone((*reinterpret_cast<std::add_pointer_t<ChangeRecord>>(_a[1]))); break;
        case 4: _t->changeRedone((*reinterpret_cast<std::add_pointer_t<ChangeRecord>>(_a[1]))); break;
        case 5: _t->dirtyStateChanged((*reinterpret_cast<std::add_pointer_t<bool>>(_a[1]))); break;
        default: ;
        }
    }
    if (_c == QMetaObject::IndexOfMethod) {
        if (QtMocHelpers::indexOfMethod<void (ChangeTracker::*)(const ChangeRecord & )>(_a, &ChangeTracker::changeRecorded, 0))
            return;
        if (QtMocHelpers::indexOfMethod<void (ChangeTracker::*)(int )>(_a, &ChangeTracker::changesApplied, 1))
            return;
        if (QtMocHelpers::indexOfMethod<void (ChangeTracker::*)(int )>(_a, &ChangeTracker::changesDiscarded, 2))
            return;
        if (QtMocHelpers::indexOfMethod<void (ChangeTracker::*)(const ChangeRecord & )>(_a, &ChangeTracker::changeUndone, 3))
            return;
        if (QtMocHelpers::indexOfMethod<void (ChangeTracker::*)(const ChangeRecord & )>(_a, &ChangeTracker::changeRedone, 4))
            return;
        if (QtMocHelpers::indexOfMethod<void (ChangeTracker::*)(bool )>(_a, &ChangeTracker::dirtyStateChanged, 5))
            return;
    }
}

const QMetaObject *tablepro::ChangeTracker::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *tablepro::ChangeTracker::qt_metacast(const char *_clname)
{
    if (!_clname) return nullptr;
    if (!strcmp(_clname, qt_staticMetaObjectStaticContent<qt_meta_tag_ZN8tablepro13ChangeTrackerE_t>.strings))
        return static_cast<void*>(this);
    return QObject::qt_metacast(_clname);
}

int tablepro::ChangeTracker::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = QObject::qt_metacall(_c, _id, _a);
    if (_id < 0)
        return _id;
    if (_c == QMetaObject::InvokeMetaMethod) {
        if (_id < 6)
            qt_static_metacall(this, _c, _id, _a);
        _id -= 6;
    }
    if (_c == QMetaObject::RegisterMethodArgumentMetaType) {
        if (_id < 6)
            *reinterpret_cast<QMetaType *>(_a[0]) = QMetaType();
        _id -= 6;
    }
    return _id;
}

// SIGNAL 0
void tablepro::ChangeTracker::changeRecorded(const ChangeRecord & _t1)
{
    QMetaObject::activate<void>(this, &staticMetaObject, 0, nullptr, _t1);
}

// SIGNAL 1
void tablepro::ChangeTracker::changesApplied(int _t1)
{
    QMetaObject::activate<void>(this, &staticMetaObject, 1, nullptr, _t1);
}

// SIGNAL 2
void tablepro::ChangeTracker::changesDiscarded(int _t1)
{
    QMetaObject::activate<void>(this, &staticMetaObject, 2, nullptr, _t1);
}

// SIGNAL 3
void tablepro::ChangeTracker::changeUndone(const ChangeRecord & _t1)
{
    QMetaObject::activate<void>(this, &staticMetaObject, 3, nullptr, _t1);
}

// SIGNAL 4
void tablepro::ChangeTracker::changeRedone(const ChangeRecord & _t1)
{
    QMetaObject::activate<void>(this, &staticMetaObject, 4, nullptr, _t1);
}

// SIGNAL 5
void tablepro::ChangeTracker::dirtyStateChanged(bool _t1)
{
    QMetaObject::activate<void>(this, &staticMetaObject, 5, nullptr, _t1);
}
QT_WARNING_POP
