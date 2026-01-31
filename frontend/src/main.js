import './style.css';
import './app.css';

import logo from './assets/images/logo-universal.png';
import { Greet } from '../wailsjs/go/main/App';
import {SelectExcelFile, ImportPurchaseVoucher} from '../wailsjs/go/main/App'


document.querySelector('#app').innerHTML = `
    <img id="logo" class="logo" style="width:100px; height:100px;">

    <div class="result" id="result">Select Excel File for Tally</div>
    
    <div class="input-box" id="input">
        <input class="input" id="filePath" type="text" readonly placeholder="No file selected..." style="width: 300px;" />
        
        <button class="btn" onclick="browseFile()">Browse</button>
    </div>

    <br>

    <button class="btn" onclick="startImport()" style="background-color: #28a745;">Import to Tally</button>
`;
document.getElementById('logo').src = logo;

// let nameElement = document.getElementById("name");
// nameElement.focus();
let resultElement = document.getElementById("result");
// Elements ko select kar lete hain
let filePathElement = document.getElementById("filePath");
let selectedPath = ""; // Yahan file ka path store hoga

// Setup the greet function
// window.greet = function () {
//     // Get name,
//     let name = nameElement.value;

//     // Check if the input is empty
//     if (name === "") return;

//     // Call App.Greet(name)
//     try {
//         Greet(name)
//             .then((result) => {
//                 // Update result with data back from App.Greet()
//                 resultElement.innerText = result;
//             })
//             .catch((err) => {
//                 console.error(err);
//             });
//     } catch (err) {
//         console.error(err);
//     }
// };

// ---------------------------------------------------
// 2. BROWSE FUNCTION (File select karne ke liye)
// ---------------------------------------------------
window.browseFile = function () {
    try {
        // Go function call: SelectExcelFile
        SelectExcelFile().then((path) => {
            // Agar user ne file select ki (cancel nahi kiya)
            if (path && path !== "") {
                selectedPath = path;
                filePathElement.value = path; // Input box me path dikhao
                resultElement.innerText = "File Selected. Click Import.";
                resultElement.style.color = "black";
            }
        });
    } catch (err) {
        console.error(err);
    }
}

// ---------------------------------------------------
// 3. IMPORT FUNCTION (Tally me upload karne ke liye)
// ---------------------------------------------------
window.startImport = function () {
    // Check karo file select hai ya nahi
    if (selectedPath === "") {
        resultElement.innerText = "Please select a file first!";
        resultElement.style.color = "red";
        return;
    }

    resultElement.innerText = "Processing... Please wait.";
    
    try {
        // Go function call: UploadToTally
        ImportPurchaseVoucher(selectedPath).then((response) => {
            // Response jo Go se aayega (Success/Fail count)
            resultElement.innerText = response;
            resultElement.style.color = "blue";
        });
    } catch (err) {
        console.error(err);
        resultElement.innerText = "Error calling backend.";
    }
};


