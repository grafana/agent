import { useState } from 'react';
import { NavLink } from 'react-router-dom';
import {
  Nav,
  Navbar as BootstrapNavbar,
  NavItem,
  NavLink as BootstrapLink,
  NavbarToggler,
  Collapse,
  UncontrolledDropdown,
  DropdownToggle,
  DropdownMenu,
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
            <DropdownMenu end>
              <NavLink to="/status/build-info" className="dropdown-item" tabIndex={0} role="menuitem">
                Runtime and build information
              </NavLink>
              <NavLink to="/status/flags" className="dropdown-item" tabIndex={0} role="menuitem">
                Command-line flags
              </NavLink>
              <NavLink to="/status/config-file" className="dropdown-item" tabIndex={0} role="menuitem">
                Configuration file
              </NavLink>
            </DropdownMenu>
          </UncontrolledDropdown>
          <NavItem>
            <BootstrapLink href="https://grafana.com/docs/agent/latest">Help</BootstrapLink>
          </NavItem>
        </Nav>
      </Collapse>
    </BootstrapNavbar>
  );
}

export default Navbar;
