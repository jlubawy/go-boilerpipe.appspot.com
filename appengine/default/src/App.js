import React, { Component } from 'react';
import {
  Alert,
  Button,
  Col,
  ControlLabel,
  Form,
  FormControl,
  FormGroup,
  Grid,
  Navbar,
  Row
} from 'react-bootstrap';


/**
 *
 */
function ErrorAlert(props) {
  if (!props.message) {
    return null;
  }

  return (
    <Alert bsStyle="danger">
      <h4>Error</h4>
      <p>{props.message}</p>
    </Alert>
  );
}


/**
 *
 */
function ExtractResults(props) {
  if (!props.doc) {
    return <p className="text-center">Try entering an article URL.</p>;
  }

  var dateStr = '';
  if (props.doc.date) {
    dateStr = (new Date(props.doc.date)).toString();
  }

  return (
    <div>
      <h1>{props.doc.title}</h1>
      {dateStr && <p><strong>Date:</strong> {dateStr}</p>}
      <p><strong>Normalized URL:</strong> {props.doc.url}</p>
      <p dangerouslySetInnerHTML={{__html: props.doc.content}}></p>
    </div>
  );
}

class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      articleURL: (localStorage) ? localStorage.getItem('articleURL') : '',
      error: null,
      results: null
    };

    this.handleChange = this.handleChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  handleChange(event) {
    event.preventDefault();
    this.setState({articleURL: event.target.value});
  }

  handleSubmit(event) {
    var self = this;
    event.preventDefault();

    if (this.state.articleURL === '') {
      this.setState({error: 'Article URL cannot be empty.'});
    } else {
      console.log('Extracting: ' + this.state.articleURL);

      // Clear previous results + errors
      this.setState({
        error: null,
        results: null
      });

      var extractURL = '/api/extract?type=html&url=' + this.state.articleURL;
      if (localStorage) {
        localStorage.setItem('articleURL', this.state.articleURL);
      }

      fetch(extractURL).then(function(resp) {
        return resp.json();
      }).then(function(json) {
        if (json.status >= 400) {
          return Promise.reject(new Error(json.message));
        } else {
          console.log(json.results);
          self.setState({results: json.results});
        }
      }).catch(function(err) {
        console.error(err);
        self.setState({error: err.message});
      });
    }
  }

  render() {
    return (
      <div>
        <Navbar>
          <Navbar.Header>
            <Navbar.Brand>
              <a href="/">Boilerpipe</a>
            </Navbar.Brand>
          </Navbar.Header>
        </Navbar>
        <Grid>
          <Row>
            <Form horizontal>
              <FormGroup>
                <Col componentClass={ControlLabel} sm={2}>
                  Article URL
                </Col>
                <Col sm={8}>
                  <FormControl type="text" value={this.state.articleURL} onChange={this.handleChange} placeholder="https://example.com/news/article" />
                </Col>
                <Col sm={2}>
                  <Button type="submit" bsStyle="primary" onClick={this.handleSubmit}>Extract</Button>
                </Col>
              </FormGroup>
            </Form>
          </Row>
          <Row>
            <Col sm={12}>
              <ErrorAlert message={this.state.error} />
              <ExtractResults doc={this.state.results} />
            </Col>
          </Row>
        </Grid>
      </div>
    );
  }
}

export default App;
