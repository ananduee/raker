import React from 'react';
import Navbar from "react-bootstrap/Navbar";
import Container from 'react-bootstrap/Container';

function Menu() {
    return (
        <Navbar variant="dark" bg="dark" expand="lg">
            <Container>
                <Navbar.Brand href="#home">Raker</Navbar.Brand>
            </Container>
        </Navbar>
    )
}

export default Menu