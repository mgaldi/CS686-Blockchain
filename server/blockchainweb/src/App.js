// import 'bootstrap/dist/css/bootstrap-theme.css';
import 'bootstrap/dist/css/bootstrap.css';
import React, { Component } from 'react';
import { Col, Row, Container } from 'react-bootstrap';
import { Notification } from 'react-notification';

import './App.css';
import Blocks from './Blocks.js';

// const DashboardPubSub = require('./pubsub').default


class App extends Component {
  constructor() {
    super()

    // State initialization
    this.state = {
      blocks: [],
      user: {},
      notification: {
        isActive: false,
        permanentNotification: false,
        message: String("")
      },
    }
  }

  toggleNotification(message) {
    this.setState({
      notification: {
        isActive: true,
        permanentNotification: false,
        message: message
      }
    })
  }


  render() {
    const notificationArea = (
      <div>
        <Notification
          isActive={this.state.notification.isActive}
          message={this.state.notification.message || ""}
          action="Dismiss"
          title="Error!"
          dismissAfter={3500}
          onDismiss={() => this.setState({ notification: { isActive: false } })}
          onClick={() => this.setState({ notification: { isActive: false } })}
        />
      </div>
    )

    return (
      <div className="App">
        {notificationArea}
        <Container fluid>
          <Row >
            <Col xs={12} md={12} lg={12}>
              <Blocks />
            </Col>
          </Row>
          </Container>
      </div>
    )
  }
}

export default App
