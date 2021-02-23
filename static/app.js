const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById('form');
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}`).then((response) => {
      response.json().then((results) => {
        Controller.updateTable(results);
      });
    });
    Controller.updateTable(["hehe"]);
  },

  updateTable: (results) => {
    const search = document.getElementById('search');
    search.classList.remove('height-100');
    search.classList.add('search-transition');
    search.classList.add('section-padding');

    const searchcol = document.getElementById('search-column');
    searchcol.classList.remove('is-6');
    searchcol.classList.add('is-8');
    searchcol.classList.add('search-column-transition');

    const table = document.getElementById('table-body');
    table.innerHTML = '';
    for (let result of results) {
      table.insertRow().innerHTML = `<td>${result}</td>`;
    }
  },
};

const form = document.getElementById('form');
form.addEventListener('submit', Controller.search);
