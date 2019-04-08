var MONTHNAMES = new Array("Jan", "Feb", "Mar", "Apr", "May", "Jun",
  "Jul", "Aug", "Sep", "Oct", "Nov", "Dec");

function formatDate(d) {
  return d.getDate() + " " + MONTHNAMES[d.getMonth()] + " " + d.getFullYear();
}

function formatURL(url) {
  var clean = url.replace("http://", "");
  clean = clean.replace("https://", "");
  clean = clean.substr(0, 50);
  return "<a href=\"" + url + "\">" + clean + "</a>";
}

function loadRedirs(howmany) {

  $('#data tr:not(:first)').remove();
  $.getJSON("/redirects/" + howmany, function(allUrls) {
    for (var i = 0; i < allUrls.length; i++) {
      var redir = allUrls[i];
      var d = new Date(redir["creation_date"] / 1000000);

      // $("#redirectTable").addRow({
      //   newRow: "<tr>" + "<td class=\"source\">" + formatURL(redir["source_url"]) + "</td>" + "<td class=\"target\">" + formatURL(redir["target_url"]) + "</td>" + "<td class=\"date\">" + formatDate(d) + "</td>" + "<td class=\"clicks\">" + redir["clicks"] + "</td>" + "</tr>",
      //   rowSpeed: 700
      // });

      $("#redirectTable").DataTable().row.add([
        formatURL(redir["source_url"]),
        formatURL(redir["target_url"]),
        formatDate(d),
        redir["clicks"]
      ]).draw();
    }
  });
}
