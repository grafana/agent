import { FC } from 'react';
import { Card, CardBody, CardText, Container } from 'reactstrap';

export enum ComponentHealth {
  HEALTHY = 'healthy',
  UNHEALTHY = 'unhealthy',
  UNKNOWN = 'unknown',
  EXITED = 'exited',
}

export interface Component {
  id: string;
  health: ComponentHealth;
}

interface ComponentListProps {
  components: Component[];
}

const ComponentList: FC<ComponentListProps> = ({ components }) => {
  return (
    <Container>
      {components.map((component) => {
        let color = 'primary';
        switch (component.health) {
          case ComponentHealth.HEALTHY:
            color = 'success';
            break;
          case ComponentHealth.UNHEALTHY:
            color = 'danger';
            break;
          case ComponentHealth.UNKNOWN:
            color = 'warning';
            break;
          case ComponentHealth.EXITED:
            color = 'primary';
            break;
        }

        return (
          <Card outline color={color} className="row">
            <CardBody>
              <CardText>
                {component.id} ({component.health})
              </CardText>
            </CardBody>
          </Card>
        );
      })}
    </Container>
  );
};

export default ComponentList;
