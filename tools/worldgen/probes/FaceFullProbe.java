package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.util.Base64;

/**
 * Dumps the exact 1.21.1 per-state classifiers vanilla's spawn column scan
 * needs beyond the motion-blocking heightmaps: whether a state's collision
 * shape fills its top face (Block.isFaceFull(shape, Direction.UP)) and whether
 * its fluid state is non-empty. Mirrors MotionBlockingProbe's output format.
 */
public final class FaceFullProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null); // Bootstrap registries.
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Method decodeState = Class.forName("dfy").getMethod("a", int.class);
        Object emptyLevel = Class.forName("dcl").getField("a").get(null); // EmptyBlockGetter.INSTANCE
        Class<?> blockPos = Class.forName("jd");
        Object zero = blockPos.getField("c").get(null); // BlockPos.ZERO
        Method collisionShape = Class.forName("dtc").getMethod("k", Class.forName("dcc"), blockPos);
        Object up = Class.forName("ji").getField("b").get(null); // Direction.UP
        Method isFaceFull = Class.forName("dfy").getMethod("a", Class.forName("exv"), Class.forName("ji"));
        Method fluidState = Class.forName("dtc").getMethod("u");
        Method fluidEmpty = Class.forName("epe").getMethod("c");
        byte[] faceFull = new byte[(26684 + 7) / 8];
        byte[] fluid = new byte[faceFull.length];
        for (int stateID = 0; stateID < 26684; stateID++) {
            Object state = decodeState.invoke(null, stateID);
            Object shape = collisionShape.invoke(state, emptyLevel, zero);
            if ((Boolean) isFaceFull.invoke(null, shape, up)) {
                faceFull[stateID / 8] |= 1 << (stateID % 8);
            }
            if (!(Boolean) fluidEmpty.invoke(fluidState.invoke(state))) {
                fluid[stateID / 8] |= 1 << (stateID % 8);
            }
        }
        System.out.println("face_full_states 26684 " + Base64.getEncoder().encodeToString(faceFull));
        System.out.println("fluid_states 26684 " + Base64.getEncoder().encodeToString(fluid));
    }
}
