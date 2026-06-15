const documentInput = document.getElementById("documentInput");
const documentLabel = document.querySelector(".document-label");
const documentIcon = document.querySelector(".document-icon");
const documentText = document.querySelector(".document-text");
const documentNameSpan = document.querySelector(".document-name");
documentInput.addEventListener("change", function () {
  if (documentInput.files.length > 0) {
    documentNameSpan.textContent = documentInput.files[0].name;
    documentText.style.display = "none";
    documentIcon.style.display = "none";
  } else {
    documentNameSpan.textContent = "";
    documentText.style.display = "inline";
  }
});
