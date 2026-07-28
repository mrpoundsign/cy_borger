var firstNames = ["SPICE", "CHROME", "RAZOR", "HEX", "VOIX", "KILOWATT", "NULL", "CYPHER", "VEX", "ASH", "CINDER", "NEXUS", "RIPPED", "GLITCH", "STATIC", "ZERO", "SHADOW", "BLADE", "VIPER", "VENOM", "HAWK", "ECHO", "GHOST", "SPECTRE"];

var toTitleCase = function(str) {
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

    function getCookieValue(name) {
        const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
        return match ? match[2] : null;
    }

    function setCookieValue(name, value) {
        document.cookie = name + "=" + value + "; path=/; max-age=31536000";
    }

    let uid = getCookieValue('cy_user_id');
    if (!uid) {
        uid = 'usr_' + Math.random().toString(36).substring(2, 11) + Date.now().toString(36);
        setCookieValue('cy_user_id', uid);
    }

    let uname = getCookieValue('cy_user_name');
    if (!uname) {
        uname = generateRandomUsername();
        setCookieValue('cy_user_name', uname);
    }
})();
