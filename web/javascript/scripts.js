// Dynamic row add
const addButton = document.querySelector(".addAHU");
const newRow = document.querySelector(".AHU_units");

function deleteRow(){
  this.parentElement.remove();
}

function addRow(){
  const AHU_name = document.createElement("td");
  AHU_name.innerHTML = '<input type ="text" name = "AHU_name"></input>';  

  const Load = document.createElement("td");
  Load.innerHTML = '<input type ="number" name = "Load" min = "0" required></input>';

  const T1cell = document.createElement("td");
  T1cell.innerHTML = '<input type ="number" name = "T1" min = "20" max = "130" required></input>';

  const T2cell = document.createElement("td");
  T2cell.innerHTML = '<input type ="number" name = "T2" min = "20" max = "130" required></input>';

  const dPahu = document.createElement("td");
  dPahu.innerHTML = '<input type ="number" name = "dPahu" min = "0" required></input>';

  const Connect_side = document.createElement("td");
  Connect_side.innerHTML = '<select type="text" name="Connection_side"><option value="left">Левый</option><option value="right">Правый</option></select>';

  const btn = document.createElement("input");
  btn.type = "button";
  btn.tagName = "delete";
  btn.value = "Удалить";
  btn.className = "delete";
  btn.innerHTML = "&times";

  btn.addEventListener("click", deleteRow);

  const Single_AHU = document.createElement("tr");
  Single_AHU.className="Single_AHU";

  newRow.appendChild(Single_AHU);
  Single_AHU.appendChild(AHU_name);
  Single_AHU.appendChild(Load);
  Single_AHU.appendChild(T1cell);
  Single_AHU.appendChild(T2cell);
  Single_AHU.appendChild(dPahu);
  Single_AHU.appendChild(Connect_side);
  Single_AHU.appendChild(btn);
}

addButton.addEventListener("click", addRow);


