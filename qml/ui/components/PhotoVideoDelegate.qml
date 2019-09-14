import QtQuick 2.4
import Ubuntu.Components 1.3
import QtGraphicalEffects 1.0

Item {
    id: thumbnailBox

    signal clicked(var mouse)

    width: {
        if (isPhotoOrVideo) {
            if (thumbnailImage.source == "") {
                height = units.gu(12);
                return units.gu(12);
            } else {
                if (thumbnailImage.isLandscape) {
                    height = units.gu(14);
                    return units.gu(26);
                } else {
                    height = units.gu(26);
                    return units.gu(14);
                }
            }
        } else {
            height = 0;
            return 0;
        }
    }

    visible: isPhotoOrVideo

    Image {
        id: thumbnailImage

        property bool isLandscape: false
        property real aspectRatio: 1

        ActivityIndicator {
            id: ai
            anchors {
                fill: parent
            }
            visible: running
            running: thumbnailImage.status != Image.Ready
        }

        anchors {
            fill: parent
        }
        asynchronous: true
        fillMode: Image.PreserveAspectFit

        sourceSize.width: 128
        sourceSize.height: 128
        source: thumbnail ? Qt.resolvedUrl(thumbnail) :  Qt.resolvedUrl(attachment)

        onStatusChanged:  {
            if (status === Image.Error) {
                width = 128
                height = 128
                source = "image://theme/image-missing"
                console.warn("Image error, source is: _" + source + "_");
            } else if (status === Image.Ready) {
                isLandscape = implicitWidth > implicitHeight
                aspectRatio = implicitWidth / implicitHeight
            }
        }
    }

    FastBlur {
        anchors.fill: thumbnailImage
        transparentBorder: true
        radius: isVideo ? 32 : 0 // Workaround for: enabled
        // enabled: isVideo // This doesn't work!
        source: thumbnailImage
        clip: true
    }

    Rectangle {
        id: videoIndicator
        anchors {
            top: thumbnailImage.top
            topMargin: units.gu(0.5)
            left: thumbnailImage.left
            leftMargin: units.gu(0.5)
        }
        width: childrenRect.width + units.gu(1)
        height: units.gu(2)
        visible: isVideo

        color: "#64000000"
        radius: 2

        Image {
            anchors {
                left: parent.left
                leftMargin: units.gu(0.5)
                verticalCenter: parent.verticalCenter
            }
            height: units.gu(1.5)
            fillMode: Image.PreserveAspectFit
            smooth: true
            antialiasing: true
            asynchronous: true
            source: Qt.resolvedUrl("../images/ic_video.png")
        }
    }

    Image {
        id: thumbnailImageOverlay
        anchors.centerIn: thumbnailImage
        width: units.gu(6)
        height: width
        visible: isVideo || (isPhoto && needsDownload)

        source: {
            if (needsDownload) {
                if (isDownloading) {
                    return Qt.resolvedUrl("../images/photocancel.png");
                } else {
                    return Qt.resolvedUrl("../images/photoload.png");
                }
            } else {
                return Qt.resolvedUrl("../images/playvideo.png");
            }
        }
    }

    MouseArea {
        anchors.fill: thumbnailImage
        onClicked: { thumbnailBox.clicked(mouse) }
    }
}
