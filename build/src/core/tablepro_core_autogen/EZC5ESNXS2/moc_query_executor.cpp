/****************************************************************************
** Meta object code from reading C++ file 'query_executor.h'
**
** Created by: The Qt Meta Object Compiler version 69 (Qt 6.10.2)
**
** WARNING! All changes made in this file will be lost!
*****************************************************************************/

#include "../../../../../include/core/query_executor.h"
#include <QtCore/qmetatype.h>

#include <QtCore/qtmochelpers.h>

#include <memory>


#include <QtCore/qxptype_traits.h>
#if !defined(Q_MOC_OUTPUT_REVISION)
#error "The header file 'query_executor.h' doesn't include <QObject>."
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
struct qt_meta_tag_ZN8tablepro13QueryExecutorE_t {};
} // unnamed namespace

template <> constexpr inline auto tablepro::QueryExecutor::qt_create_metaobjectdata<qt_meta_tag_ZN8tablepro13QueryExecutorE_t>()
{
    namespace QMC = QtMocConstants;
    QtMocHelpers::StringRefStorage qt_stringData {
        "tablepro::QueryExecutor",
        "queryStarted",
        "",
        "query",
        "queryFinished",
        "QueryResult",
        "result",
        "queryError",
        "error",
        "transactionStarted",
        "transactionCommitted",
        "transactionRolledBack"
    };

    QtMocHelpers::UintData qt_methods {
        // Signal 'queryStarted'
        QtMocHelpers::SignalData<void(const QString &)>(1, 2, QMC::AccessPublic, QMetaType::Void, {{
            { QMetaType::QString, 3 },
        }}),
        // Signal 'queryFinished'
        QtMocHelpers::SignalData<void(const QString &, const QueryResult &)>(4, 2, QMC::AccessPublic, QMetaType::Void, {{
            { QMetaType::QString, 3 }, { 0x80000000 | 5, 6 },
        }}),
        // Signal 'queryError'
        QtMocHelpers::SignalData<void(const QString &, const QString &)>(7, 2, QMC::AccessPublic, QMetaType::Void, {{
            { QMetaType::QString, 3 }, { QMetaType::QString, 8 },
        }}),
        // Signal 'transactionStarted'
        QtMocHelpers::SignalData<void()>(9, 2, QMC::AccessPublic, QMetaType::Void),
        // Signal 'transactionCommitted'
        QtMocHelpers::SignalData<void()>(10, 2, QMC::AccessPublic, QMetaType::Void),
        // Signal 'transactionRolledBack'
        QtMocHelpers::SignalData<void()>(11, 2, QMC::AccessPublic, QMetaType::Void),
    };
    QtMocHelpers::UintData qt_properties {
    };
    QtMocHelpers::UintData qt_enums {
    };
    return QtMocHelpers::metaObjectData<QueryExecutor, qt_meta_tag_ZN8tablepro13QueryExecutorE_t>(QMC::MetaObjectFlag{}, qt_stringData,
            qt_methods, qt_properties, qt_enums);
}
Q_CONSTINIT const QMetaObject tablepro::QueryExecutor::staticMetaObject = { {
    QMetaObject::SuperData::link<QObject::staticMetaObject>(),
    qt_staticMetaObjectStaticContent<qt_meta_tag_ZN8tablepro13QueryExecutorE_t>.stringdata,
    qt_staticMetaObjectStaticContent<qt_meta_tag_ZN8tablepro13QueryExecutorE_t>.data,
    qt_static_metacall,
    nullptr,
    qt_staticMetaObjectRelocatingContent<qt_meta_tag_ZN8tablepro13QueryExecutorE_t>.metaTypes,
    nullptr
} };

void tablepro::QueryExecutor::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    auto *_t = static_cast<QueryExecutor *>(_o);
    if (_c == QMetaObject::InvokeMetaMethod) {
        switch (_id) {
        case 0: _t->queryStarted((*reinterpret_cast<std::add_pointer_t<QString>>(_a[1]))); break;
        case 1: _t->queryFinished((*reinterpret_cast<std::add_pointer_t<QString>>(_a[1])),(*reinterpret_cast<std::add_pointer_t<QueryResult>>(_a[2]))); break;
        case 2: _t->queryError((*reinterpret_cast<std::add_pointer_t<QString>>(_a[1])),(*reinterpret_cast<std::add_pointer_t<QString>>(_a[2]))); break;
        case 3: _t->transactionStarted(); break;
        case 4: _t->transactionCommitted(); break;
        case 5: _t->transactionRolledBack(); break;
        default: ;
        }
    }
    if (_c == QMetaObject::IndexOfMethod) {
        if (QtMocHelpers::indexOfMethod<void (QueryExecutor::*)(const QString & )>(_a, &QueryExecutor::queryStarted, 0))
            return;
        if (QtMocHelpers::indexOfMethod<void (QueryExecutor::*)(const QString & , const QueryResult & )>(_a, &QueryExecutor::queryFinished, 1))
            return;
        if (QtMocHelpers::indexOfMethod<void (QueryExecutor::*)(const QString & , const QString & )>(_a, &QueryExecutor::queryError, 2))
            return;
        if (QtMocHelpers::indexOfMethod<void (QueryExecutor::*)()>(_a, &QueryExecutor::transactionStarted, 3))
            return;
        if (QtMocHelpers::indexOfMethod<void (QueryExecutor::*)()>(_a, &QueryExecutor::transactionCommitted, 4))
            return;
        if (QtMocHelpers::indexOfMethod<void (QueryExecutor::*)()>(_a, &QueryExecutor::transactionRolledBack, 5))
            return;
    }
}

const QMetaObject *tablepro::QueryExecutor::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *tablepro::QueryExecutor::qt_metacast(const char *_clname)
{
    if (!_clname) return nullptr;
    if (!strcmp(_clname, qt_staticMetaObjectStaticContent<qt_meta_tag_ZN8tablepro13QueryExecutorE_t>.strings))
        return static_cast<void*>(this);
    return QObject::qt_metacast(_clname);
}

int tablepro::QueryExecutor::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
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
void tablepro::QueryExecutor::queryStarted(const QString & _t1)
{
    QMetaObject::activate<void>(this, &staticMetaObject, 0, nullptr, _t1);
}

// SIGNAL 1
void tablepro::QueryExecutor::queryFinished(const QString & _t1, const QueryResult & _t2)
{
    QMetaObject::activate<void>(this, &staticMetaObject, 1, nullptr, _t1, _t2);
}

// SIGNAL 2
void tablepro::QueryExecutor::queryError(const QString & _t1, const QString & _t2)
{
    QMetaObject::activate<void>(this, &staticMetaObject, 2, nullptr, _t1, _t2);
}

// SIGNAL 3
void tablepro::QueryExecutor::transactionStarted()
{
    QMetaObject::activate(this, &staticMetaObject, 3, nullptr);
}

// SIGNAL 4
void tablepro::QueryExecutor::transactionCommitted()
{
    QMetaObject::activate(this, &staticMetaObject, 4, nullptr);
}

// SIGNAL 5
void tablepro::QueryExecutor::transactionRolledBack()
{
    QMetaObject::activate(this, &staticMetaObject, 5, nullptr);
}
QT_WARNING_POP
