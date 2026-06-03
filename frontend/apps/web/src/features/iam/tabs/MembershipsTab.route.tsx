import { Outlet } from "react-router-dom";
import MembershipsTab from "./MembershipsTab";

export function Component() {
  return (
    <>
      <MembershipsTab />
      <Outlet />
    </>
  );
}
