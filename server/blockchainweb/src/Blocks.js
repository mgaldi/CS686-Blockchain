import React, { Component } from 'react';
import { Button, Row, Col, Form, FormLabel, Container } from 'react-bootstrap';
import 'react-bootstrap-table-next/dist/react-bootstrap-table2.css';
import ToolkitProvider, { ColumnToggle, Search, CSVExport } from 'react-bootstrap-table2-toolkit';
import BootstrapTable from 'react-bootstrap-table-next';

const { ToggleList } = ColumnToggle;
const { SearchBar, ClearSearchButton } = Search;
const { ExportCSVButton } = CSVExport;
class Blocks extends Component {
    constructor() {
        super()
        this.state = {
            title : '',
            author:'',
            pages: '',
            year: '',
            blocks: [],
            blocks: [],
            columns: [{
                dataField: 'Height',
                text: 'height',
                style: { overflowWrap: "break-word" },
                sort: true
            }, {
                dataField: 'Timestamp',
                text: 'Timestamp',
                style: { overflowWrap: "break-word" },
                sort: true
            }, {
                dataField: 'ParentHash',
                text: 'Parent Hash',
                style: { overflowWrap: "break-word" },
                sort: false
            }, {
                dataField: 'RootHash',
                text: 'RootHash',
                style: { overflowWrap: "break-word" },
                sort: false
            }, {
                dataField: 'Data',
                text: 'Data',
                style: { overflowWrap: "break-word" },
                sort: false,
                formatter: (cell) => {
                
                    
                        return <>{
                            <div> 
                            {cell && cell.map((book, index) => 
                            <div>
                                <div className="text-left">
                                    <span className="font-weight-bold text-center">Book {index}</span>
                                    </div>
                                    <div>
                                <p><b>Title:</b> {book.Title}</p>
                                <p><b>Author:</b> {book.Author}</p>
                                <p><b>Pages:</b> {book.Pages}</p>
                                <p><b>Year:</b> {book.Year}</p></div>
                                </div>
                                )}                            
                
                         </div>}</>
                }
            }]
        }
        this.title = ""
        this.author = ""
        this.pages = ""
        this.year = ""
        this.remote = this.remote.bind(this)
        this.newBook = this.newBook.bind(this)
        this.doQuery = this.doQuery.bind(this)
        this.handleInputChange = this.handleInputChange.bind(this);

    }
    // this.updateData = this.updateData.bind(this)

    remote(remoteObj){
        remoteObj.insertRow = true
        return remoteObj
    }
    handleInputChange(event) {
        const name = event.target.name;
    
        this.setState({
            [name]: event.target.type === 'number' ? parseInt(event.target.value) : event.target.value
        });
      }
      
   async newBook(event){
    
        event.preventDefault()

        try {
            let data = {
                Title : this.state.title,
                Author : this.state.author,
                Pages : parseInt(this.state.pages),
                Year: parseInt(this.state.year)
            }
            console.log(JSON.stringify(data))
            let response = await fetch("/new", {
                method: "POST",
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data),
            }).finally(response => {
                console.log("Done")
            })
            this.doQuery()

            
        }catch (e){
            console.log(e)
        }       
    }

    async doQuery(){
        var ports= new Array()
        try {
            let response = await fetch("http://localhost:8080/validators", {
              method: "GET",
              // headers: {
              //   'Accept': 'application/json',
              //   'Content-Type': 'application/json',
              // },
            })
            response = await response.json()
            for(var i = 0; i < response.length; i++) {
                var obj = response[i]
                ports.push((obj.ProcessNr * 4) + 46000)
                console.log(obj.ProcessNr)
            }
            console.log(JSON.stringify(response))
          } catch (e) {
            console.log(e)
          }
          var data = this
          var rootHashes = new Array()
          var heights = new Array()
          var ok = true
          for(var i = 0; i < ports.length; i++){

            try {
                var req = `http://localhost:${ports[i]}/list`
                let response = await fetch(req, {
                  method: "GET",
                })
                data = await response.json()
              
                var obj = data[data.length - 1]
                rootHashes.push(obj.RootHash)
                heights.push(obj.Height)
                console.log(obj.Height)
                console.log(JSON.stringify(data))
              } catch (e) {
                console.log(e)
              }
              if(heights.length > 1){
        
                  if(heights[heights.length - 1] !== heights[heights.length - 2]){
                      ok = false
                  }
              }
              if(rootHashes.length > 1){
                if(rootHashes[rootHashes.length - 1] !== rootHashes[rootHashes.length - 2]){
                    ok = false
                }
            }

          }
          if(ok){
            this.resultOk(data)
          }
    }
    async componentDidMount() { 
        console.log("Query for blocks.")
        this.doQuery()
      }

      async resultOk(response) {
        if (response.error) {
          console.log("error")
        } else {
          this.setState({
            blocks: response
          })
        }
      }
    
    render() {
        const expandRow={
            renderer: row => (
            <div>
                <p>Parent Hash: {row.ParentHash}</p>
                <p>Root Hash: {row.RootHash}</p>
            </div>
            )
        };

        return (
            <div className="bootstrap-table-margin fix-height mt-5" >
                
                <Form onSubmit={this.newBook} className="mx-5">
                    
                        
                    <Form.Group controlId="form.Book" className="text-left">
                    <Form.Row>
                        
                        <Col>
                            <Form.Label>Title</Form.Label>
                            <Form.Control type="text" placeholder="Title" onChange={this.handleInputChange}  name="title" value={this.state.title}/>
                            <br />
                            <Form.Label>Author</Form.Label>
                            <Form.Control type="text" placeholder="Author" onChange={this.handleInputChange}  name="author" value={this.state.author}/>
                            <br />
                        </Col>
                        <Col>
                            <Form.Label>Number of Pages</Form.Label>
                            <Form.Control type="text" placeholder="Pages" onChange={this.handleInputChange}  name="pages" value={this.state.pages}/>
                            <br />
                            <Form.Label>Year</Form.Label>
                            <Form.Control type="text" placeholder="Year" onChange={this.handleInputChange}  name="year" value={this.state.year}/>
                        </Col>
                        </Form.Row>

                    </Form.Group>
                   
            
                    <Button variant="primary" type="submit" >
                        Submit
                    </Button>
                </Form>

               
                <ToolkitProvider
                    
                    keyField="Height"
                    data={this.state.blocks}
                    columns={this.state.columns}
                    search
                    bootstrap4
                    columnToggle    
                
                >
                    {
                        props => (
                            <div>
                                <Row className="mt-5">
                                    <Col lg={6}>
                                        <SearchBar {...props.searchProps} />
                                        <ClearSearchButton {...props.searchProps} />

                                    </Col>
                                    <Col lg={6}>
                                        <ToggleList {...props.columnToggleProps} />
                                    </Col>
                                </Row>



                                <hr />
                                <BootstrapTable
                                    expandRow={ expandRow }
                                    {...props.baseProps}
                                />
                                {/* <ExportCSVButton {...props.csvProps}><Button>Export CSV!!</Button></ExportCSVButton> */}
                            </div>
                        )
                    }

                </ToolkitProvider>
            </div>
        )
    }
}

export default Blocks
