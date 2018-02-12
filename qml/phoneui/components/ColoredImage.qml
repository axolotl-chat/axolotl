/*
 * Copyright (C) 2014 Canonical, Ltd.
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

import QtQuick 2.3

Item {
    id: root

    /*!
       \qmlproperty color color
    */
    property alias color: colorizedImage.keyColorOut

    /*!
       \qmlproperty color keyColor
    */
    property alias keyColor: colorizedImage.keyColorIn

    property alias mirror: image.mirror
    property alias source: image.source
    property alias asynchronous: image.asynchronous

    Image {
        id: image

        anchors.fill: parent
        visible: false
    }

    ShaderEffect {
        id: colorizedImage

        anchors.fill: parent
        visible: active

        // Whether or not a color has been set.
        property bool active: keyColorOut != Qt.rgba(0.0, 0.0, 0.0, 0.0)

        property Image source: active && image.status === Image.Ready ? image : null
        property color keyColorOut: Qt.rgba(0.0, 0.0, 0.0, 0.0)
        property color keyColorIn: "#ffffff"
        property real threshold: 0.1

        readonly property string mirroredShader: "
            uniform highp mat4 qt_Matrix;
            attribute highp vec4 qt_Vertex;
            attribute highp vec2 qt_MultiTexCoord0;
            varying highp vec2 qt_TexCoord0;
            void main() {
                qt_TexCoord0 = vec2(1.0 - qt_MultiTexCoord0.s, qt_MultiTexCoord0.t);
                gl_Position = qt_Matrix * qt_Vertex;
            }"

        readonly property string originalShader: "
            uniform highp mat4 qt_Matrix;
            attribute highp vec4 qt_Vertex;
            attribute highp vec2 qt_MultiTexCoord0;
            varying highp vec2 qt_TexCoord0;
            void main() {
                qt_TexCoord0 = qt_MultiTexCoord0;
                gl_Position = qt_Matrix * qt_Vertex;
            }"

        vertexShader: root.mirror ? mirroredShader : originalShader

        fragmentShader: "
            varying highp vec2 qt_TexCoord0;
            uniform sampler2D source;
            uniform highp vec4 keyColorOut;
            uniform highp vec4 keyColorIn;
            uniform lowp float threshold;
            uniform lowp float qt_Opacity;
            void main() {
                lowp vec4 sourceColor = texture2D(source, qt_TexCoord0);
                gl_FragColor = mix(vec4(keyColorOut.rgb, 1.0) * sourceColor.a,
                               sourceColor,
                               step(threshold, distance(sourceColor.rgb / sourceColor.a, keyColorIn.rgb))) * qt_Opacity;
            }"
    }
}
