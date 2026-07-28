// The single JS entry point. Vite bundles this — and the CSS it imports — into
// web/static/. These libraries used to be <script> tags in layout.templ; now they
// are npm dependencies compiled into one bundle, so the browser makes one request
// for JS and one for CSS, from our own server.
import './app.css'

import htmx from 'htmx.org'
import Alpine from 'alpinejs'

// Exposed on window for console access and browser extensions. htmx wires up its
// DOM processing on import; Alpine needs an explicit start().
window.htmx = htmx
window.Alpine = Alpine
Alpine.start()
