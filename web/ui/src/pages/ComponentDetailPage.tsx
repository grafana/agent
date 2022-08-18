import { FC } from 'react';
import { ComponentView } from '../features/component/ComponentView';
import { componentInfoByID } from '../features/component/types';
import { testComponentDetail, testComponents } from '../testdata/data';

export const ComponentDetailPage: FC = () => {
  const infoByID = componentInfoByID(testComponents);
  return <ComponentView component={testComponentDetail} info={infoByID} />;
};
