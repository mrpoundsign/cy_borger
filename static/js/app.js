const firstNames = ["SPICE", "CHROME", "RAZOR", "HEX", "VOIX", "KILOWATT", "NULL", "CYPHER", "VEX", "ASH", "CINDER", "NEXUS", "RIPPED", "GLITCH", "STATIC", "ZERO", "SHADOW", "BLADE", "VIPER", "VENOM", "HAWK", "ECHO", "GHOST", "SPECTRE"];

function toTitleCase(str) {
    return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
}

function generateRandomUsername() {
    const name1 = toTitleCase(firstNames[Math.floor(Math.random() * firstNames.length)]);
    let name2 = toTitleCase(firstNames[Math.floor(Math.random() * firstNames.length)]);
    while (name2 === name1) {
        name2 = toTitleCase(firstNames[Math.floor(Math.random() * firstNames.length)]);
    }
    const num = Math.floor(1000 + Math.random() * 9000);
    return name1 + name2 + "#" + num;
}

function randomizeThemeColor() {
    const colors = ['#ffe600', '#ff0055', '#00f0ff', '#00ff66', '#9d00ff', '#ff5500'];
    let current = localStorage.getItem('cy_theme_accent') || '#ffe600';
    let next = colors[Math.floor(Math.random() * colors.length)];
    while (next === current) {
        next = colors[Math.floor(Math.random() * colors.length)];
    }
    localStorage.setItem('cy_theme_accent', next);
    document.documentElement.style.setProperty('--accent-color', next);
    document.documentElement.style.setProperty('--border-color', next);
}

function copyText(text, btn) {
    navigator.clipboard.writeText(text).then(() => {
        const orig = btn.innerText;
        btn.innerText = "COPIED!";
        btn.style.borderColor = "#00ff66";
        btn.style.color = "#00ff66";
        setTimeout(() => {
            btn.innerText = orig;
            btn.style.borderColor = "";
            btn.style.color = "";
        }, 2000);
    });
}

(function () {
    const accent = localStorage.getItem('cy_theme_accent') || '#ffe600';
    document.documentElement.style.setProperty('--accent-color', accent);
    document.documentElement.style.setProperty('--border-color', accent);

    let uid = localStorage.getItem('cy_user_id');
    if (!uid) {
        const match = document.cookie.match(/(?:^|; )cy_user_id=([^;]*)/);
        if (match && match[1]) {
            uid = match[1];
        } else {
            uid = 'usr_' + Math.random().toString(36).substring(2, 11) + Date.now().toString(36);
        }
        localStorage.setItem('cy_user_id', uid);
    }
    document.cookie = "cy_user_id=" + uid + "; path=/; max-age=31536000";

    let uname = localStorage.getItem('cy_user_name');
    if (!uname) {
        uname = generateRandomUsername();
        localStorage.setItem('cy_user_name', uname);
    }

    window.cyStore = {
        getUserId: function () { return localStorage.getItem('cy_user_id'); },
        getUserName: function () { return localStorage.getItem('cy_user_name'); },
        setUserName: function (name) { localStorage.setItem('cy_user_name', name); },
        hasEditedName: function () { return localStorage.getItem('cy_user_name_edited') === 'true'; },
        setEditedName: function () { localStorage.setItem('cy_user_name_edited', 'true'); },
        saveGMGame: function (gameId, gmCode, name) {
            let games = JSON.parse(localStorage.getItem('cy_gm_games') || '{}');
            games[gameId] = { id: gameId, gmCode: gmCode, name: name, timestamp: Date.now() };
            localStorage.setItem('cy_gm_games', JSON.stringify(games));
        },
        getGMGames: function () { return JSON.parse(localStorage.getItem('cy_gm_games') || '{}'); },
        saveCharEdit: function (charId, editCode, name) {
            let chars = JSON.parse(localStorage.getItem('cy_user_chars') || '{}');
            chars[charId] = { id: charId, editCode: editCode, name: name, timestamp: Date.now() };
            localStorage.setItem('cy_user_chars', JSON.stringify(chars));
        },
        getCharEdits: function () { return JSON.parse(localStorage.getItem('cy_user_chars') || '{}'); }
    };
})();
