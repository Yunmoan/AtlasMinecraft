package atlasprobe;

// Tests the vanilla CubicSpline float arithmetic independently of noise.
public final class SplineProbe {
    public static void main(String[] args) throws Exception {
        Class<?> spline = Class.forName("axp");
        Class<?> coordinate = Class.forName("azq");
        Object identity = coordinate.getField("a").get(null);
        Object builder = spline.getMethod("a", coordinate).invoke(null, identity);
        var add = builder.getClass().getMethod("a", float.class, float.class, float.class);
        add.invoke(builder, -1.1f, 0.044f, 0.0f);
        add.invoke(builder, -0.5f, -0.22f, 0.4f);
        add.invoke(builder, 0.5f, 0.3f, -0.3f);
        add.invoke(builder, 1.0f, 0.69f, 0.4f);
        Object built = builder.getClass().getMethod("a").invoke(builder);
        float[] inputs = {-2f, -1.1f, -0.8f, -0.5f, 0f, 0.5f, 0.75f, 1f, 2f};
        for (int i = 0; i < inputs.length; i++) {
            float value = (Float) spline.getMethod("a", Object.class).invoke(built, inputs[i]);
            System.out.printf("spline %d %s%n", i, Integer.toUnsignedString(Float.floatToRawIntBits(value)));
        }
    }
}
