// FUNCTION DRAG OVER
$("body").on("dragover", function (e) {
  e.preventDefault()
  $("body").addClass("drag-over")
});

// FUNCTION DRAG LEAVE
$("body").on("dragleave", function () {
  $("body").removeClass("drag-over")
});

// FUNCTION UPLOAD IMAGE MULTIPLE DROP
$("body").on("drop", function (e) {
  e.preventDefault();
  $("body").removeClass("drag-over");
  const files = e.originalEvent.dataTransfer.files;
  if (files.length > 0) {
    $('#listProgressUpload').show()
    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      const fileName = file.name
      if (fileName.endsWith(".PNG") || fileName.endsWith(".png") || fileName.endsWith(".JPEG") || fileName.endsWith(".jpeg") || fileName.endsWith(".JPG") || fileName.endsWith(".jpg")) {
        const listItem = `<li><i class="fa fa-image"></i> <h1>${fileName}</h1> <div class="progressBar"><span id="progress-${i}-percent" class="progressPercent"><i class="fas fa-spinner fa-spin" style="margin-right: 5px;"></i><strong id="countProgressPercent"></strong></span><progress id="progress-${i}" value='0' max='100'></progress></div></li>`;
        $('#fileNamesList').append(listItem)
        // Add a progress bar for each file
        const progressBar = "progress-" + i
        // Upload the file and update the progress bar
        setTimeout(function(){
          uploadImage(file, progressBar)
        }, 1000);
      } else {
        $('#fileNamesList').append(`<li><i class="fa fa-times"></i> <h1>Skipped an invalid file: ` + fileName + `</h1>`)
      }
    }
  } else {
    $('#fileNamesList').append(`<li><i class="fa fa-times"></i> <h1>Please drop one or more files</h1></li>`)
  }
})


// FUNCTION MINIMIZE SECTION UPLOAD LIST PROGRESS UPLOAD
function minimizeListProgress(el) {
  $('.fileNamesContainer').toggle()
  $(el).find('i').toggleClass("fa fa-chevron-down fa fa-chevron-up");
}

// FUNCTION CLOSE SECTION LIST PROGRESS UPLOAD
function closeListProgress() {
  $('#fileNamesList').empty()
  $('#fileInput').val("")
  $('#listProgressUpload').hide()
}

// FUNCTION UPLOAD FILE OR PICTURE
function uploadImage(file, progressBar) {
  const formData = new FormData();
  formData.append("file", file);
  
  $.ajax({
    url: "/compress", // Replace with your server endpoint
    type: "POST",
    data: formData,
    processData: false,
    contentType: false,
    xhr: function () {
      const xhr = new window.XMLHttpRequest();
      xhr.upload.addEventListener("progress", function (e) {
        if (e.lengthComputable) {
          const percentComplete = (e.loaded / e.total) * 100;
          $("#" + progressBar).val(percentComplete);
          // Update the tooltip's content with the progress percentage
          $("#" + progressBar + "-percent #countProgressPercent").text(percentComplete.toFixed(0) + "%")
        }
      })
      return xhr;
    },
    success: function (response) {
      // Handle the response from the server if needed
      if (response == "OK") {
        $("#" + progressBar + "-percent").find('i').removeClass().addClass('fa fa-check-circle checkSuccess')
      } else if (response == "NG") {
        $("#" + progressBar + "-percent").find('i').removeClass().addClass('fa fa-times checkError')
      }
      // // REFRESH DATA TABLE
      // loadImage()
    },
    error: function (xhr, status, error) {
      console.error(xhr.responseText);
    },
  });
}