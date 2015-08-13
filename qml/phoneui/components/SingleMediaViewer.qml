/*
 * Copyright 2015 Canonical Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import QtQuick 2.2
import QtMultimedia 5.0
import Ubuntu.Components 1.0

Item {
    id: viewer
    property bool pinchInProgress: zoomPinchArea.active
    property var mediaSource
    property real maxDimension
    property bool showThumbnail: true

    property bool isVideo: videoPreviewSource !== ""
    property bool userInteracting: pinchInProgress || flickable.sizeScale != 1.0
    property bool fullyZoomed: flickable.sizeScale == zoomPinchArea.maximumZoom
    property bool fullyUnzoomed: flickable.sizeScale == zoomPinchArea.minimumZoom

    property alias paintedHeight: image.paintedHeight
    property alias paintedWidth: image.paintedWidth

    signal clicked()

    function zoomIn(centerX, centerY, factor) {
        flickable.scaleCenterX = centerX / (flickable.sizeScale * flickable.width);
        flickable.scaleCenterY = centerY / (flickable.sizeScale * flickable.height);
        flickable.sizeScale = factor;
    }

    function zoomOut() {
        if (flickable.sizeScale != 1.0) {
            flickable.scaleCenterX = flickable.contentX / flickable.width / (flickable.sizeScale - 1);
            flickable.scaleCenterY = flickable.contentY / flickable.height / (flickable.sizeScale - 1);
            flickable.sizeScale = 1.0;
        }
    }

    function reset() {
        if (viewer.isVideo) {
            // Intentionally pausing, not stopping, as resuming is faster.
            videoPreview.pause();
        } else {
            zoomOut();
        }
    }

    ActivityIndicator {
        id: activityIndicator
        anchors.centerIn: parent
        visible: running && !viewer.isVideo
        running: image.status != Image.Ready
    }

    PinchArea {
        id: zoomPinchArea
        anchors.fill: parent

        property real initialZoom
        property real maximumScale: 3.0
        property real minimumZoom: 1.0
        property real maximumZoom: 3.0
        property bool active: false
        property var center

        onPinchStarted: {
            if (viewer.isVideo) return;

            active = true;
            initialZoom = flickable.sizeScale;
            center = zoomPinchArea.mapToItem(media, pinch.startCenter.x, pinch.startCenter.y);
            zoomIn(center.x, center.y, initialZoom);
        }
        onPinchUpdated: {
            if (viewer.isVideo) return;

            var zoomFactor = MathUtils.clamp(initialZoom * pinch.scale, minimumZoom, maximumZoom);
            flickable.sizeScale = zoomFactor;
        }
        onPinchFinished: {
            active = false;
        }

        Flickable {
            id: flickable
            anchors.fill: parent
            contentWidth: media.width
            contentHeight: media.height
            contentX: (sizeScale - 1) * scaleCenterX * width
            contentY: (sizeScale - 1) * scaleCenterY * height
            interactive: !viewer.pinchInProgress

            property real sizeScale: 1.0
            property real scaleCenterX: 0.0
            property real scaleCenterY: 0.0
            Behavior on sizeScale {
                enabled: !viewer.pinchInProgress
                UbuntuNumberAnimation {duration: UbuntuAnimation.FastDuration}
            }
            Behavior on scaleCenterX {
                UbuntuNumberAnimation {duration: UbuntuAnimation.FastDuration}
            }
            Behavior on scaleCenterY {
                UbuntuNumberAnimation {duration: UbuntuAnimation.FastDuration}
            }

            Item {
                id: media

                width: flickable.width * flickable.sizeScale
                height: flickable.height * flickable.sizeScale

                Image {
                    id: image
                    anchors.fill: parent
                    asynchronous: true
                    cache: false
                    source: photoPreviewSource
                    sourceSize {
                        width: viewer.maxDimension
                        height: viewer.maxDimension
                    }
                    fillMode: Image.PreserveAspectFit
                    visible: viewer.showThumbnail
                    opacity: status == Image.Ready ? 1.0 : 0.0
                    Behavior on opacity { UbuntuNumberAnimation {duration: UbuntuAnimation.FastDuration} }

                }

                Image {
                    id: highResolutionImage
                    anchors.fill: parent
                    asynchronous: true
                    cache: false
                    // Load image using the GalleryStandardImageProvider to ensure EXIF orientation
                    source: flickable.sizeScale > 1.0 ? photoPreviewSource : ""
                    sourceSize {
                        width: width
                        height: height
                    }
                    fillMode: Image.PreserveAspectFit
                }

                MediaPlayer {
                    id: videoPreview

                    property bool isPlaying: playbackState === MediaPlayer.PlayingState

                    autoLoad: true
                    autoPlay: true
                    source: videoPreviewSource
                }

                VideoOutput {
                    id: videoOutput
                    anchors.fill: parent
                    visible: videoPreviewSource !== ""
                    source: videoPreview
                }
            }

            Icon {
                width: units.gu(5)
                height: units.gu(5)
                anchors.centerIn: parent
                name: "media-playback-start"
                color: "white"
                opacity: 0.8
                visible: viewer.isVideo && !videoPreview.isPlaying && !activityIndicator.visible
            }

            MouseArea {
                anchors.fill: parent
                onDoubleClicked: {
                    clickTimer.stop();

                    if (viewer.isVideo) {
                        // Rewind.
                        videoPreview.stop();
                        videoPreview.play();
                        return;
                    }
                    if (flickable.sizeScale < zoomPinchArea.maximumZoom) {
                        zoomIn(mouse.x, mouse.y, zoomPinchArea.maximumZoom);
                    } else {
                        zoomOut();
                    }
                }
                onClicked: {
                    clickTimer.start();
                }

                Timer {
                    id: clickTimer
                    interval: 150
                    onTriggered: {
                        if (viewer.isVideo) {
                            videoPreview.isPlaying ? videoPreview.pause() : videoPreview.play();
                        }
                        viewer.clicked();
                    }
                }
            }
        }
    }
}
