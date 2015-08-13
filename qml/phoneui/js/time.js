.pragma library

var MILLIS_IN_DAY = 1000 * 60 * 60 * 24;
var MILLIS_IN_WEEK = 7 * MILLIS_IN_DAY;

function format(i18n, time) {
    var date = new Date(time);
    var today = new Date();
    // TRANSLATORS: localized time string, see available formats at
    // http://doc.qt.io/qt-5/qml-qtqml-qt.html#formatDateTime-method
    var timeFormatted = Qt.formatTime(date, i18n.tr("hh:mm"));
    // TRANSLATORS: localized date string, see available formats at
    // http://doc.qt.io/qt-5/qml-qtqml-qt.html#formatDateTime-method
    var dateFormatted = Qt.formatDate(date, i18n.tr("MMM d"));

    var isToday = date.getDate() === today.getDate();
    if (isToday) {
        return timeFormatted;
    } else {
        var oneWeekAgo = new Date(today.getTime() - MILLIS_IN_WEEK);
        var isLessThanWeek =
                date.getTime() > oneWeekAgo.getTime() &&
                date.getDate() !== oneWeekAgo.getDate();
        if (isLessThanWeek) {
            return Qt.formatDate(date, "ddd");
        }
    }
    return dateFormatted;
}

function formatSection(i18n, time) {
    var date = new Date(time);
    // TRANSLATORS: localized date string, see available formats at
    // http://doc.qt.io/qt-5/qml-qtqml-qt.html#formatDateTime-method
    var dateFormatted = Qt.formatDate(date, i18n.tr("MMM d"));

    return dateFormatted;
}

function formatLastSeen(i18n, time) {
    var date = new Date(time);
    var today = new Date();
    var yesterday = new Date(today - MILLIS_IN_DAY);
    // TRANSLATORS: localized time string, see available formats at
    // http://doc.qt.io/qt-5/qml-qtqml-qt.html#formatDateTime-method
    var timeFormatted = Qt.formatTime(date, i18n.tr("hh:mm"));
    // TRANSLATORS: localized date string, see available formats at
    // http://doc.qt.io/qt-5/qml-qtqml-qt.html#formatDateTime-method
    var dateFormatted = Qt.formatDate(date, i18n.tr("MMM d"));

    var isToday = date.getDate() === today.getDate();
    if (isToday) {
        // TRANSLATORS %1 refers to a time of the day
        return i18n.tr("today at %1").arg(timeFormatted);
    } else if (date.getDate() === yesterday.getDate()) {
        // TRANSLATORS %1 refers to a time of the day
        return i18n.tr("yesterday at %1").arg(timeFormatted);
    } else {
        // TRANSLATORS %1 refers to a date, and %1 to a time of the day
        return i18n.tr("%1 at %2").arg(dateFormatted).arg(timeFormatted);
    }
}

function formatTimeOnly(i18n, time) {
    var date = new Date(time);
    // TRANSLATORS: localized time string, see available formats at
    // http://doc.qt.io/qt-5/qml-qtqml-qt.html#formatDateTime-method
    var timeFormatted = Qt.formatTime(date, i18n.tr("hh:mm"));

    return timeFormatted;
}

function areSameDay(date1, date2) {
    var firstDate = new Date(date1)
    var secondDate = new Date(date2)
    if (!firstDate || !secondDate)
        return false
    return firstDate.getFullYear() == secondDate.getFullYear()
            && firstDate.getMonth() == secondDate.getMonth()
            && firstDate.getDate() == secondDate.getDate()
}
