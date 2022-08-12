import { useState } from 'react';
import { NavLink } from 'react-router-dom';
import {
  Nav,
  Navbar as BootstrapNavbar,
  NavItem,
  NavbarToggler,
  Collapse,
  UncontrolledDropdown,
  DropdownToggle,
  DropdownMenu,
  DropdownItem,
} from 'reactstrap';
import logo from '../images/logo.svg';

function Navbar() {
  const [collapsed, setCollapsed] = useState(true);

  const toggleNavbar = () => setCollapsed(!collapsed);

  return (
    <BootstrapNavbar className="app-navbar" expand="lg">
      <NavLink to="/" className="navbar-brand" style={{ height: '45px' }}>
        <img
          src={logo}
          alt="Grafana Agent Logo"
          title="Grafana Agent"
          style={{
            height: '25px',
          }}
        />
      </NavLink>
      <NavbarToggler onClick={toggleNavbar} />
      <Collapse isOpen={!collapsed} navbar>
        <Nav navbar>
          <NavItem>
            <NavLink to="/dag" className="nav-link">
              DAG
            </NavLink>
          </NavItem>
          <UncontrolledDropdown nav inNavbar>
            <DropdownToggle nav caret>
              Status
            </DropdownToggle>
            <DropdownMenu right>
              <DropdownItem>Runtime and build information</DropdownItem>
              <DropdownItem>Command-line flags</DropdownItem>
              <DropdownItem>Configuration file</DropdownItem>
            </DropdownMenu>
          </UncontrolledDropdown>
        </Nav>
      </Collapse>
    </BootstrapNavbar>
  );
}

export default Navbar;
