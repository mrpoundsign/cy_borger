const fs = require('fs');
const { JSDOM } = require('jsdom');

const html = fs.readFileSync('templates/game.html', 'utf8');
const dom = new JSDOM(html);
const document = dom.window.document;

const container = document.getElementById('party-container');
console.log("Container exists:", !!container);
if (container) {
    console.log("Container parent:", container.parentElement.tagName);
    console.log("Container next sibling:", container.nextElementSibling ? container.nextElementSibling.tagName : 'none');
    const inner = document.getElementById('party-container-inner');
    console.log("Inner exists:", !!inner);
    if (inner) {
        console.log("Inner is child of container:", inner.parentElement === container);
        console.log("Inner children count:", inner.children.length);
        console.log("Inner next sibling:", inner.nextElementSibling ? inner.nextElementSibling.tagName : 'none');
    }
}
