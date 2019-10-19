.pragma library

var AVATARS = [
    ["#8179d7", "_violet.png"],
    ["#f2749a", "_pink.png"  ],
    ["#7ec455", "_green.png" ],
    ["#f3c34a", "_yellow.png"],
    ["#5b9dd8", "_blue.png"  ],
    ["#62b8cd", "_aqua.png"  ],
    ["#ed8b4a", "_orange.png"],
    ["#d95848", "_red.png"   ]
]

function getColor(userId) {
    return AVATARS[userId % 8][0];
}

function getAvatar(userId, isGroup) {
    isGroup = isGroup || false;
    return (isGroup ? "group" : "user") + AVATARS[userId % 8][1];
}

