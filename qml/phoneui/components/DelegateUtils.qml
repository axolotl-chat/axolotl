import QtQuick 2.4
import Ubuntu.Content 1.1

Item {
	function getMediaTypeString(mediaType) {
		switch (mediaType) {
			case ContentType.Pictures:
			return "Photo"
			case ContentType.Videos:
			return "Video"
			case ContentType.Music:
			return "Audio"
			default:
			return "Text";
		}
	}

    function humanFileSize(bytes, si) {
        var thresh = si ? 1000 : 1024;
        if (bytes < thresh) return bytes + ' B';
        var units = si ? ['kB','MB','GB','TB','PB','EB','ZB','YB'] : ['KiB','MiB','GiB','TiB','PiB','EiB','ZiB','YiB'];
        var u = -1;
        do {
            bytes /= thresh;
            ++u;
        } while(bytes >= thresh);
        return bytes.toFixed(1) + ' ' + units[u];
    }
}
