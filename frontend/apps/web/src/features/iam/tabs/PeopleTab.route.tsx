import { Outlet } from "react-router-dom";
import PeopleTab from "./PeopleTab";

export function Component() {
  return (
    <>
      <PeopleTab />
      <Outlet />
    </>
  );
}
