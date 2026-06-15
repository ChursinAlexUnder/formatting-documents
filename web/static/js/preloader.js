document.getElementById("document-download-form").addEventListener("submit", function() {
    document.getElementById("preloader").style.display = "block";
});
window.addEventListener("pageshow", function() {
    document.getElementById("preloader").style.display = "none";
});
