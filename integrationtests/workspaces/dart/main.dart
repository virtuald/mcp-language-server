// Main test file for Dart definition and hover tests
import 'helper.dart';
import 'types.dart';

void main() {
  print('Hello, World!');

  var helper = HelperClass('test');
  helper.process();

  SharedClass shared = SharedClass(42);
  print(shared.getValue());

  // Use types
  Color color = Color.red;
  print(color);

  // Use constant
  print(SHARED_CONSTANT);

  // Use function
  var newHelper = createHelper('new');
  print(newHelper.getName());
}
