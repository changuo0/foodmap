import logo from './logo.svg';
import './App.css';

function makeRequest(dbOrCal,params,callback) {
  const req = new XMLHttpRequest()
  req.addEventListener("load", () => {
    callback(JSON.parse(req.responseText))
  })
  let url = "/" + dbOrCal + "/"
  for (var k in params)
    url += k + "=" + params[k] + "/"
  req.open("GET", url)
  req.send()
}

function App() {
  makeRequest("db",{"key1":"val1","key2":"val2"}, (x) => alert("db req returned "+JSON.stringify(x)))
  makeRequest("cal",{"calkey1":"calval1","calkey2":"calval2"}, (x) => alert("cal req returned "+JSON.stringify(x)))
  return (
    <div className="App">
      <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Edit <code>src/App.js</code> and save to reload.
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
      </header>
    </div>
  );
}

export default App;
