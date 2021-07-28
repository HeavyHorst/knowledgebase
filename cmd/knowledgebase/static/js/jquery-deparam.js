/*
 TERMS OF USE - EASING EQUATIONS
 Open source under the MIT License.
 Copyright (c) 2013 AceMetrix
 github.com/AceMetrix/jquery-deparam
*/
(function(e){if("function"===typeof require&&"object"===typeof exports&&"object"===typeof module){try{var h=require("jquery")}catch(l){}module.exports=e(h)}else if("function"===typeof define&&define.amd)define(["jquery"],function(b){return e(b)});else{var b;try{b=(0,eval)("this")}catch(l){b=window}b.deparam=e(b.jQuery)}})(function(e){var h=function(b,e){var f={},h={"true":!0,"false":!1,"null":null};if(!b)return f;b.replace(/\+/g," ").split("&").forEach(function(c){var a=c.split("=");c=decodeURIComponent(a[0]);
var b=f,k=0,d=c.split("]["),g=d.length-1;/\[/.test(d[0])&&/\]$/.test(d[g])?(d[g]=d[g].replace(/\]$/,""),d=d.shift().split("[").concat(d),g=d.length-1):g=0;if(2===a.length)if(a=decodeURIComponent(a[1]),e&&(a=a&&!isNaN(a)&&+a+""===a?+a:"undefined"===a?void 0:void 0!==h[a]?h[a]:a),g)for(;k<=g;k++)c=""===d[k]?b.length:d[k],b=b[c]=k<g?b[c]||(d[k+1]&&isNaN(d[k+1])?{}:[]):a;else"[object Array]"===Object.prototype.toString.call(f[c])?f[c].push(a):{}.hasOwnProperty.call(f,c)?f[c]=[f[c],a]:f[c]=a;else c&&(f[c]=
e?void 0:"")});return f};e&&(e.prototype.deparam=e.deparam=h);return h});