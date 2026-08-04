import { UNAUTHED_ERR_MSG } from '@shared/const';
import { initTRPC, TRPCError } from "@trpc/server";
import superjson from "superjson";
import type { TrpcContext } from "./context";

const t = initTRPC.context<TrpcContext>().create({
  transformer: superjson,
});

export const router = t.router;
export const publicProcedure = t.procedure;

const requireUser = t.middleware(async opts => {
  const { ctx, next } = opts;

  if (!ctx.user) {
    throw new TRPCError({ code: "UNAUTHORIZED", message: UNAUTHED_ERR_MSG });
  }

  return next({
    ctx: {
      ...ctx,
      user: ctx.user,
    },
  });
});

export const protectedProcedure = t.procedure.use(requireUser);

export type NodeScopeRole = "viewer" | "operator" | "administrator";

function nodeScopeRoleForTemplateUser(role: string | null | undefined): NodeScopeRole {
  // The development template only has user/admin. Production NodeScope maps
  // Supabase roles directly; this preserves the same least-privilege shape in
  // the console workspace without granting operator privileges by default.
  return role === "admin" ? "administrator" : "viewer";
}

function roleRank(role: NodeScopeRole): number {
  switch (role) {
    case "viewer":
      return 1;
    case "operator":
      return 2;
    case "administrator":
      return 3;
  }
}

function requireNodeScopeRole(minimumRole: NodeScopeRole) {
  return t.middleware(async opts => {
    const { ctx, next } = opts;
    if (!ctx.user) {
      throw new TRPCError({ code: "UNAUTHORIZED", message: UNAUTHED_ERR_MSG });
    }

    const actualRole = nodeScopeRoleForTemplateUser(ctx.user.role);
    if (roleRank(actualRole) < roleRank(minimumRole)) {
      throw new TRPCError({
        code: "FORBIDDEN",
        message: `NodeScope ${minimumRole} role is required`,
      });
    }

    return next({
      ctx: {
        ...ctx,
        user: ctx.user,
        nodeScopeRole: actualRole,
      },
    });
  });
}

export const viewerProcedure = t.procedure.use(requireNodeScopeRole("viewer"));
export const operatorProcedure = t.procedure.use(requireNodeScopeRole("operator"));
export const administratorProcedure = t.procedure.use(requireNodeScopeRole("administrator"));

// Backward-compatible alias retained for the starter system router. New
// NodeScope procedures should use administratorProcedure by name.
export const adminProcedure = administratorProcedure;
