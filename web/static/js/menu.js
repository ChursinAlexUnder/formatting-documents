const buttons = document.querySelectorAll(".window-one-button");

buttons.forEach(button => {
  button.addEventListener("click", function (event) {
    event.preventDefault();
    buttons.forEach(b => b.classList.remove("active"));
    button.classList.add("active");
  });
});
